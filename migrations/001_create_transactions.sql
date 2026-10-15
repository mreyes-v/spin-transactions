CREATE TABLE IF NOT EXISTS transactions (
  id UUID PRIMARY KEY,
  account_id TEXT NOT NULL,
  type TEXT NOT NULL,
  amount NUMERIC(18,2) NOT NULL,
  currency TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT NOT NULL,
  provider_transaction_id TEXT,
  balance_after NUMERIC(18,2),
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions (account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions (status);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions (type);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions (created_at DESC);
