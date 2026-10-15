package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"spin-transactions/internal/application"
	"spin-transactions/internal/domain/transaction"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

// TransactionRepository serves both services. Write and read share this table.
type TransactionRepository struct {
	db *sql.DB
}

var (
	_ application.TransactionWriter = (*TransactionRepository)(nil)
	_ application.TransactionReader = (*TransactionRepository)(nil)
)

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Save(ctx context.Context, tx transaction.Transaction) error {
	var providerID any
	if tx.ProviderTransactionID != "" {
		providerID = tx.ProviderTransactionID
	}
	var balanceAfter any
	if tx.Status == transaction.StatusExecuted {
		balanceAfter = tx.BalanceAfter
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO transactions (
			id, account_id, type, amount, currency, description,
			status, provider_transaction_id, balance_after, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		tx.ID, tx.AccountID, string(tx.Type), tx.Amount, tx.Currency, tx.Description,
		string(tx.Status), providerID, balanceAfter, tx.CreatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) List(ctx context.Context, filter transaction.ListFilter) ([]transaction.Transaction, int, error) {
	page, limit := normalizePagination(filter.Page, filter.Limit)
	offset := (page - 1) * limit
	where, args := buildListWhere(filter)

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM transactions"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	listArgs := append(append([]any{}, args...), limit, offset)
	listQuery := `
		SELECT id, account_id, type, amount, currency, description,
		       status, provider_transaction_id, balance_after, created_at
		FROM transactions` + where + `
		ORDER BY created_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	items := make([]transaction.Transaction, 0)
	for rows.Next() {
		var tx transaction.Transaction
		var providerID sql.NullString
		var balanceAfter sql.NullFloat64
		var txType, status string
		if err := rows.Scan(
			&tx.ID, &tx.AccountID, &txType, &tx.Amount, &tx.Currency, &tx.Description,
			&status, &providerID, &balanceAfter, &tx.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}
		tx.Type = transaction.Type(txType)
		tx.Status = transaction.Status(status)
		if providerID.Valid {
			tx.ProviderTransactionID = providerID.String
		}
		if balanceAfter.Valid {
			tx.BalanceAfter = balanceAfter.Float64
		}
		items = append(items, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate transactions: %w", err)
	}
	return items, total, nil
}

func (r *TransactionRepository) CurrentBalance(ctx context.Context, accountID string) (float64, error) {
	const query = `
		SELECT COALESCE(SUM(
			CASE
				WHEN type = 'CREDIT' THEN amount
				WHEN type = 'DEBIT' THEN -amount
				ELSE 0
			END
		), 0)
		FROM transactions
		WHERE account_id = $1 AND status = $2
	`
	var balance float64
	if err := r.db.QueryRowContext(ctx, query, accountID, string(transaction.StatusExecuted)).Scan(&balance); err != nil {
		return 0, fmt.Errorf("sum executed balance: %w", err)
	}
	return balance, nil
}

func buildListWhere(filter transaction.ListFilter) (string, []any) {
	conditions := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if filter.AccountID != "" {
		args = append(args, filter.AccountID)
		conditions = append(conditions, fmt.Sprintf("account_id = $%d", len(args)))
	}
	if filter.Status != "" {
		args = append(args, string(filter.Status))
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if filter.Type != "" {
		args = append(args, string(filter.Type))
		conditions = append(conditions, fmt.Sprintf("type = $%d", len(args)))
	}
	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}
