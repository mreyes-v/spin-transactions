package application

import (
	"context"
	"fmt"
	"log"

	"spin-transactions/internal/domain/account"
	tx "spin-transactions/internal/domain/transaction"
)

// WriteService executes a transaction: validate, call the provider, apply the aggregate, persist.
type WriteService struct {
	provider  PaymentProvider
	repo      TransactionWriter
	publisher EventPublisher
}

func NewWriteService(provider PaymentProvider, repo TransactionWriter, publisher EventPublisher) *WriteService {
	return &WriteService{provider: provider, repo: repo, publisher: publisher}
}

func (s *WriteService) Create(ctx context.Context, cmd tx.CreateTransaction) (tx.Transaction, error) {
	if err := cmd.Validate(); err != nil {
		return tx.Transaction{}, err
	}

	result, err := s.provider.Execute(ctx, tx.ProviderExecuteRequest{
		AccountID: cmd.AccountID,
		Type:      cmd.Type,
		Amount:    cmd.Amount,
		Currency:  cmd.Currency,
	})
	if err != nil {
		return tx.Transaction{}, err
	}

	balance, err := s.repo.CurrentBalance(ctx, cmd.AccountID)
	if err != nil {
		return tx.Transaction{}, fmt.Errorf("compute balance: %w", err)
	}

	agg := account.New(cmd.AccountID, balance)
	out, event, err := agg.ApplyProviderResult(cmd, result)
	if err != nil {
		return tx.Transaction{}, err
	}
	if err := s.repo.Save(ctx, out); err != nil {
		return tx.Transaction{}, fmt.Errorf("persist transaction: %w", err)
	}
	if err := s.publisher.Publish(ctx, event); err != nil {
		log.Printf("error publish domain event: %v", err)
	}

	if out.Status != tx.StatusExecuted {
		return tx.Transaction{}, &tx.ProviderRejectionError{
			Transaction: out,
			Code:        result.Code,
			Message:     result.Message,
		}
	}
	return out, nil
}
