-- +goose Up
CREATE TABLE subscriptions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     text NOT NULL,
    plan        text NOT NULL,
    status      text NOT NULL DEFAULT 'active' CONSTRAINT chk_subscriptions_status
        CHECK (status IN ('active', 'expired', 'cancelled')),
    started_at  timestamptz NOT NULL DEFAULT now(),
    expires_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_subscriptions_user_created ON subscriptions (user_id, created_at DESC);
CREATE INDEX idx_subscriptions_user_status ON subscriptions (user_id, status);

-- +goose Down
DROP TABLE IF EXISTS subscriptions;
