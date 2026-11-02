-- +goose Up
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
