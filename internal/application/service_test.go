package application_test

import (
	"context"
	"errors"
	"testing"

	"spin-transactions/internal/application"
	"spin-transactions/internal/domain/events"
	tx "spin-transactions/internal/domain/transaction"
)

type stubProvider struct{ res tx.ProviderResult }

func (s *stubProvider) Execute(_ context.Context, _ tx.ProviderExecuteRequest) (tx.ProviderResult, error) {
	return s.res, nil
}

type stubRepo struct {
	bal   float64
	saved []tx.Transaction
	items []tx.Transaction
	total int
}

func (s *stubRepo) Save(_ context.Context, t tx.Transaction) error {
	s.saved = append(s.saved, t)
	return nil
}

func (s *stubRepo) List(_ context.Context, f tx.ListFilter) ([]tx.Transaction, int, error) {
	if len(s.items) == 0 {
		return []tx.Transaction{{ID: "1", AccountID: f.AccountID}}, 1, nil
	}
	return s.items, s.total, nil
}

func (s *stubRepo) CurrentBalance(_ context.Context, _ string) (float64, error) {
	return s.bal, nil
}

type stubPub struct{ names []string }

func (s *stubPub) Publish(_ context.Context, e events.DomainEvent) error {
	s.names = append(s.names, e.Name())
	return nil
}

func TestCreate_ApprovedPublishesEvent(t *testing.T) {
	pub := &stubPub{}
	repo := &stubRepo{bal: 10}
	svc := application.NewWriteService(
		&stubProvider{res: tx.ProviderResult{TransactionID: "p", Status: tx.ProviderStatusApproved}},
		repo,
		pub,
	)
	got, err := svc.Create(context.Background(), tx.CreateTransaction{
		AccountID: "a", Type: tx.TypeCredit, Amount: 5, Currency: "MXN", Description: "d",
	})
	if err != nil || got.BalanceAfter != 15 || len(pub.names) != 1 || pub.names[0] != "TransactionExecuted" || len(repo.saved) != 1 {
		t.Fatalf("got=%+v err=%v events=%v saved=%d", got, err, pub.names, len(repo.saved))
	}
}

func TestCreate_RejectedPersistsAndPublishes(t *testing.T) {
	pub := &stubPub{}
	repo := &stubRepo{}
	svc := application.NewWriteService(
		&stubProvider{res: tx.ProviderResult{Status: tx.ProviderStatusRejected, Code: "C", Message: "m"}},
		repo,
		pub,
	)
	_, err := svc.Create(context.Background(), tx.CreateTransaction{
		AccountID: "a", Type: tx.TypeCredit, Amount: 5, Currency: "MXN", Description: "d",
	})
	var rej *tx.ProviderRejectionError
	if !errors.As(err, &rej) || pub.names[0] != "TransactionRejected" || len(repo.saved) != 1 {
		t.Fatalf("err=%v events=%v saved=%d", err, pub.names, len(repo.saved))
	}
}

func TestRead_List(t *testing.T) {
	svc := application.NewReadService(&stubRepo{})
	items, total, err := svc.List(context.Background(), tx.ListFilter{AccountID: "acc"})
	if err != nil || total != 1 || items[0].AccountID != "acc" {
		t.Fatalf("%v %d %+v", err, total, items)
	}
}
