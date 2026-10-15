package events

import "time"

// DomainEvent is something that already happened inside the aggregate.
type DomainEvent interface {
	Name() string
	OccurredAt() time.Time
}

type TransactionExecuted struct {
	TransactionID string
	AccountID     string
	At            time.Time
}

func (e TransactionExecuted) Name() string          { return "TransactionExecuted" }
func (e TransactionExecuted) OccurredAt() time.Time { return e.At }

type TransactionRejected struct {
	TransactionID string
	AccountID     string
	Code          string
	At            time.Time
}

func (e TransactionRejected) Name() string          { return "TransactionRejected" }
func (e TransactionRejected) OccurredAt() time.Time { return e.At }
