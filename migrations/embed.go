package migrations

import _ "embed"

//go:embed 001_create_transactions.sql
var CreateTransactionsSQL string
