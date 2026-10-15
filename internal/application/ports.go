package application

import (
	"context"

	"spin-transactions/internal/domain/events"
	tx "spin-transactions/internal/domain/transaction"
)

// PaymentProvider is the driven port toward the external PSP.
// The domain speaks ProviderResult; the HTTP adapter translates the wire format.
type PaymentProvider interface {
	Execute(ctx context.Context, req tx.ProviderExecuteRequest) (tx.ProviderResult, error)
}

// TransactionWriter is the persistence port used by the write service.
type TransactionWriter interface {
	Save(ctx context.Context, t tx.Transaction) error
	CurrentBalance(ctx context.Context, accountID string) (float64, error)
}

// TransactionReader is the persistence port used by the read service.
type TransactionReader interface {
	List(ctx context.Context, filter tx.ListFilter) ([]tx.Transaction, int, error)
}

// EventPublisher delivers domain events raised by the aggregate.
type EventPublisher interface {
	Publish(ctx context.Context, event events.DomainEvent) error
}
