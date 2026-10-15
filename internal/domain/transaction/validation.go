package transaction

import "fmt"

const (
	MinAmount         = 1.00
	MaxDebitAmount    = 10000.00
	SupportedCurrency = "MXN"
)

// Validate is the domain policy for an execution request.
// It does not know HTTP status codes or the payment provider.
func (c CreateTransaction) Validate() error {
	if c.Type != TypeCredit && c.Type != TypeDebit {
		return &ValidationError{Message: fmt.Sprintf("type must be one of: %s, %s", TypeCredit, TypeDebit)}
	}
	if c.Currency != SupportedCurrency {
		return &ValidationError{Message: fmt.Sprintf("only %s currency is supported", SupportedCurrency)}
	}
	if c.Amount <= MinAmount {
		return &ValidationError{Message: fmt.Sprintf("amount must be greater than %.2f", MinAmount)}
	}
	if c.Type == TypeDebit && c.Amount > MaxDebitAmount {
		return &ValidationError{Message: fmt.Sprintf("DEBIT amount cannot exceed %.2f", MaxDebitAmount)}
	}
	return nil
}
