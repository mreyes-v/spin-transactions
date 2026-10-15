package httpapi

import (
	"time"

	tx "spin-transactions/internal/domain/transaction"
)

type transactionResponse struct {
	ID                    string  `json:"id"`
	AccountID             string  `json:"accountId"`
	Type                  string  `json:"type"`
	Amount                float64 `json:"amount"`
	Currency              string  `json:"currency"`
	Description           string  `json:"description"`
	Status                string  `json:"status"`
	ProviderTransactionID string  `json:"providerTransactionId"`
	BalanceAfter          float64 `json:"balanceAfter"`
	CreatedAt             string  `json:"createdAt"`
}

func toResp(t tx.Transaction) transactionResponse {
	createdAt := ""
	if !t.CreatedAt.IsZero() {
		createdAt = t.CreatedAt.Format(time.RFC3339)
	}
	return transactionResponse{
		ID:                    t.ID,
		AccountID:             t.AccountID,
		Type:                  string(t.Type),
		Amount:                t.Amount,
		Currency:              t.Currency,
		Description:           t.Description,
		Status:                string(t.Status),
		ProviderTransactionID: t.ProviderTransactionID,
		BalanceAfter:          t.BalanceAfter,
		CreatedAt:             createdAt,
	}
}
