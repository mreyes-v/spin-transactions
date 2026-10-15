package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"spin-transactions/internal/domain/transaction"
)

func TestHTTPClient_Execute_Approved(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/provider/v1/execute" || r.Header.Get("X-API-Key") != "demo-api-key" {
			t.Fatalf("path=%s key=%s", r.URL.Path, r.Header.Get("X-API-Key"))
		}
		var body executeRequestBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.AccountID != "acc-123456" || body.Type != "CREDIT" || body.Amount != 1500 || body.Currency != "MXN" {
			t.Fatalf("body=%+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"transactionId": "txn-789",
			"status": "APPROVED",
			"balance": 5500.00,
			"executedAt": "2025-03-15T10:30:00Z"
		}`))
	}))
	defer server.Close()

	got, err := NewHTTPClient(server.URL, "demo-api-key").Execute(context.Background(), transaction.ProviderExecuteRequest{
		AccountID: "acc-123456",
		Type:      transaction.TypeCredit,
		Amount:    1500,
		Currency:  "MXN",
	})
	if err != nil {
		t.Fatal(err)
	}
	wantAt := time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC)
	if got.TransactionID != "txn-789" || got.Status != transaction.ProviderStatusApproved || got.Balance != 5500 || !got.ExecutedAt.Equal(wantAt) {
		t.Fatalf("%+v", got)
	}
}

func TestHTTPClient_Execute_Rejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"status":"REJECTED","code":"INSUFFICIENT_FUNDS","message":"no funds"}`))
	}))
	defer server.Close()

	got, err := NewHTTPClient(server.URL, "demo-api-key").Execute(context.Background(), transaction.ProviderExecuteRequest{
		AccountID: "acc",
		Type:      transaction.TypeDebit,
		Amount:    10,
		Currency:  "MXN",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != transaction.ProviderStatusRejected || got.Code != "INSUFFICIENT_FUNDS" {
		t.Fatalf("%+v", got)
	}
}
