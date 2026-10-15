package account

import (
	"testing"

	"spin-transactions/internal/domain/events"
	tx "spin-transactions/internal/domain/transaction"
)

func TestAccount_ApplyProviderResult_ApprovedCredit(t *testing.T) {
	agg := New("acc", 100)
	out, event, err := agg.ApplyProviderResult(
		tx.CreateTransaction{AccountID: "acc", Type: tx.TypeCredit, Amount: 40, Currency: "MXN", Description: "d"},
		tx.ProviderResult{TransactionID: "p1", Status: tx.ProviderStatusApproved},
	)
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != tx.StatusExecuted || out.BalanceAfter != 140 || agg.Balance != 140 {
		t.Fatalf("out=%+v agg.Balance=%v", out, agg.Balance)
	}
	executed, ok := event.(events.TransactionExecuted)
	if !ok || executed.AccountID != "acc" || executed.Name() != "TransactionExecuted" {
		t.Fatalf("event=%T %+v", event, event)
	}
}

func TestAccount_ApplyProviderResult_ApprovedDebit(t *testing.T) {
	agg := New("acc", 100)
	out, _, err := agg.ApplyProviderResult(
		tx.CreateTransaction{AccountID: "acc", Type: tx.TypeDebit, Amount: 25, Currency: "MXN", Description: "d"},
		tx.ProviderResult{TransactionID: "p2", Status: tx.ProviderStatusApproved},
	)
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != tx.StatusExecuted || out.BalanceAfter != 75 || agg.Balance != 75 {
		t.Fatalf("out=%+v agg.Balance=%v", out, agg.Balance)
	}
}

func TestAccount_ApplyProviderResult_RejectedDoesNotChangeBalance(t *testing.T) {
	agg := New("acc", 100)
	out, event, err := agg.ApplyProviderResult(
		tx.CreateTransaction{AccountID: "acc", Type: tx.TypeDebit, Amount: 40, Currency: "MXN", Description: "d"},
		tx.ProviderResult{Status: tx.ProviderStatusRejected, Code: "X"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != tx.StatusRejected || agg.Balance != 100 || out.BalanceAfter != 0 {
		t.Fatalf("out=%+v balance=%v", out, agg.Balance)
	}
	rejected, ok := event.(events.TransactionRejected)
	if !ok || rejected.Code != "X" || rejected.Name() != "TransactionRejected" {
		t.Fatalf("event=%T %+v", event, event)
	}
}
