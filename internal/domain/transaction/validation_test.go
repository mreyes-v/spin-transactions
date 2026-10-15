package transaction

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	cmd := CreateTransaction{AccountID: "a", Type: TypeCredit, Amount: 10, Currency: SupportedCurrency, Description: "d"}
	if err := cmd.Validate(); err != nil {
		t.Fatal(err)
	}
	cmd.Currency = "USD"
	err := cmd.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("got %T", err)
	}
}

func TestValidate_DebitAboveMax(t *testing.T) {
	cmd := CreateTransaction{AccountID: "a", Type: TypeDebit, Amount: MaxDebitAmount + 1, Currency: SupportedCurrency, Description: "d"}
	err := cmd.Validate()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("got %v", err)
	}
}
