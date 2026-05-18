CREATE TABLE IF NOT EXISTS expenses (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    amount NUMERIC(10, 2) NOT NULL CHECK (amount > 0),
    description TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_expenses_user_time ON expenses(user_id, created_at DESC);