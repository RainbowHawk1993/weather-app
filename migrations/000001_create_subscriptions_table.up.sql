CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    frequency VARCHAR(10) NOT NULL CHECK (frequency IN ('hourly', 'daily')),
    confirmation_token VARCHAR(36) UNIQUE,
    is_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    unsubscribe_token VARCHAR(36) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (email, city)
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_confirmation_token ON subscriptions (confirmation_token);
CREATE INDEX IF NOT EXISTS idx_subscriptions_unsubscribe_token ON subscriptions (unsubscribe_token);
