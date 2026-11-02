-- +goose Up
CREATE TABLE subscriptions (
 duration_months integer NOT NULL DEFAULT 1 CHECK (duration_months > 0),
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

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
