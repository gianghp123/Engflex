-- +goose Up
CREATE TABLE subscriptions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     text NOT NULL,
    plan        text NOT NULL,
    status      text NOT NULL DEFAULT 'pending' CONSTRAINT chk_subscriptions_status
        CHECK (status IN ('pending', 'active', 'expired', 'cancelled', 'failed')),
    amount      bigint NOT NULL,
    currency    text NOT NULL DEFAULT 'VND',
    started_at  timestamptz,
    expires_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_subscriptions_user_created ON subscriptions (user_id, created_at DESC);
CREATE INDEX idx_subscriptions_user_status ON subscriptions (user_id, status);

CREATE TABLE payment_transactions (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id   uuid NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    provider          text NOT NULL DEFAULT 'vnpay',
    provider_txn_ref  text NOT NULL,
    provider_txn_no   text,
    amount            bigint NOT NULL,
    status            text NOT NULL DEFAULT 'pending' CONSTRAINT chk_payment_transactions_status
        CHECK (status IN ('pending', 'succeeded', 'failed')),
    response_code     text,
    raw_response      jsonb,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_txn_ref)
);
CREATE INDEX idx_payment_transactions_subscription_id ON payment_transactions (subscription_id);

-- +goose Down
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS subscriptions;
