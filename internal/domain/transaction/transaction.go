package transaction

import "time"

type Type string

const (
	TypeCredit Type = "CREDIT"
	TypeDebit  Type = "DEBIT"
)

type Status string

const (
	StatusExecuted Status = "EXECUTED"
	StatusRejected Status = "REJECTED"
)

// Transaction is the entity produced by the Account aggregate.
type Transaction struct {
	ID                    string
	AccountID             string
	Type                  Type
	Amount                float64
	Currency              string
	Description           string
	Status                Status
	ProviderTransactionID string
	BalanceAfter          float64
	CreatedAt             time.Time
}

// CreateTransaction is the input of the execute use case, expressed in domain language.
type CreateTransaction struct {
	AccountID   string
	Type        Type
	Amount      float64
	Currency    string
	Description string
}

type ListFilter struct {
	AccountID string
	Status    Status
	Type      Type
	Page      int
	Limit     int
}

type ProviderStatus string

const (
	ProviderStatusApproved ProviderStatus = "APPROVED"
	ProviderStatusRejected ProviderStatus = "REJECTED"
)

type ProviderExecuteRequest struct {
	AccountID string
	Type      Type
	Amount    float64
	Currency  string
}

type ProviderResult struct {
	TransactionID string
	Status        ProviderStatus
	Balance       float64
	ExecutedAt    time.Time
	Code          string
	Message       string
}

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }

type ProviderRejectionError struct {
	Transaction Transaction
	Code        string
	Message     string
}

func (e *ProviderRejectionError) Error() string { return e.Message }
