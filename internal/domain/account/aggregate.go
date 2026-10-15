package account

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"spin-transactions/internal/domain/events"
	tx "spin-transactions/internal/domain/transaction"
)

// Account is the aggregate root. It owns the balance invariant and decides
// which domain event follows a provider result.
type Account struct {
	ID      string
	Balance float64
}

func New(id string, currentBalance float64) *Account {
	return &Account{ID: id, Balance: currentBalance}
}

// ApplyProviderResult records the transaction and, when the provider approved it,
// mutates the balance. The event is a domain decision; application only publishes it.
func (a *Account) ApplyProviderResult(
	cmd tx.CreateTransaction,
	result tx.ProviderResult,
) (tx.Transaction, events.DomainEvent, error) {
	now := time.Now().UTC()
	out := tx.Transaction{
		ID:          uuid.NewString(),
		AccountID:   cmd.AccountID,
		Type:        cmd.Type,
		Amount:      cmd.Amount,
		Currency:    cmd.Currency,
		Description: cmd.Description,
		CreatedAt:   now,
	}
	switch result.Status {
	case tx.ProviderStatusApproved:
		out.Status = tx.StatusExecuted
		out.ProviderTransactionID = result.TransactionID
		switch cmd.Type {
		case tx.TypeCredit:
			a.Balance += cmd.Amount
		case tx.TypeDebit:
			a.Balance -= cmd.Amount
		default:
			return tx.Transaction{}, nil, fmt.Errorf("unsupported type: %s", cmd.Type)
		}
		out.BalanceAfter = a.Balance
		return out, events.TransactionExecuted{
			TransactionID: out.ID,
			AccountID:     out.AccountID,
			At:            now,
		}, nil
	case tx.ProviderStatusRejected:
		out.Status = tx.StatusRejected
		return out, events.TransactionRejected{
			TransactionID: out.ID,
			AccountID:     out.AccountID,
			Code:          result.Code,
			At:            now,
		}, nil
	default:
		return tx.Transaction{}, nil, fmt.Errorf("unsupported provider status: %s", result.Status)
	}
}
