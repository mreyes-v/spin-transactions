package application

import (
	"context"

	tx "spin-transactions/internal/domain/transaction"
)

// ReadService lists transactions from the same table the write service persists.
type ReadService struct {
	repo TransactionReader
}

func NewReadService(repo TransactionReader) *ReadService {
	return &ReadService{repo: repo}
}

func (s *ReadService) List(ctx context.Context, filter tx.ListFilter) ([]tx.Transaction, int, error) {
	return s.repo.List(ctx, filter)
}
