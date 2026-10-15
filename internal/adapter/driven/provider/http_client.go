package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"spin-transactions/internal/application"
	"spin-transactions/internal/domain/transaction"
)

var _ application.PaymentProvider = (*HTTPClient)(nil)

type HTTPClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewHTTPClient(baseURL, apiKey string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type executeRequestBody struct {
	AccountID string  `json:"accountId"`
	Type      string  `json:"type"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type approvedResponseBody struct {
	TransactionID string  `json:"transactionId"`
	Status        string  `json:"status"`
	Balance       float64 `json:"balance"`
	ExecutedAt    string  `json:"executedAt"`
}

type rejectedResponseBody struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *HTTPClient) Execute(ctx context.Context, req transaction.ProviderExecuteRequest) (transaction.ProviderResult, error) {
	body, err := json.Marshal(executeRequestBody{
		AccountID: req.AccountID,
		Type:      string(req.Type),
		Amount:    req.Amount,
		Currency:  req.Currency,
	})
	if err != nil {
		return transaction.ProviderResult{}, fmt.Errorf("error in marshal provider request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/provider/v1/execute", bytes.NewReader(body))
	if err != nil {
		return transaction.ProviderResult{}, fmt.Errorf("error in create provider request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return transaction.ProviderResult{}, fmt.Errorf("error in call provider: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return transaction.ProviderResult{}, fmt.Errorf("error in read provider response: %w", err)
	}

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		var approved approvedResponseBody
		if err := json.Unmarshal(respBody, &approved); err != nil {
			return transaction.ProviderResult{}, fmt.Errorf("error in decode approved provider response: %w", err)
		}
		executedAt, err := time.Parse(time.RFC3339, approved.ExecutedAt)
		if err != nil {
			return transaction.ProviderResult{}, fmt.Errorf("error in parse executedAt: %w", err)
		}
		return transaction.ProviderResult{
			TransactionID: approved.TransactionID,
			Status:        transaction.ProviderStatus(approved.Status),
			Balance:       approved.Balance,
			ExecutedAt:    executedAt,
		}, nil
	}

	if resp.StatusCode >= 400 {
		var rejected rejectedResponseBody
		if err := json.Unmarshal(respBody, &rejected); err != nil {
			return transaction.ProviderResult{}, fmt.Errorf("error in decode rejected provider response: %w", err)
		}
		return transaction.ProviderResult{
			Status:  transaction.ProviderStatus(rejected.Status),
			Code:    rejected.Code,
			Message: rejected.Message,
		}, nil
	}

	return transaction.ProviderResult{}, fmt.Errorf("unexpected error in provider status %d: %s", resp.StatusCode, string(respBody))
}
