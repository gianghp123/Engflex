-- +goose Up
-- Metered state: used + reserved per (user, counter, period). This table
-- is the authoritative snapshot; usage_events is the audit ledger. The
-- CHECKs make a negative counter unrepresentable.
CREATE TABLE usage_counters (
    user_id      text        NOT NULL,
    counter_key  text        NOT NULL REFERENCES premium_features(key),
    period_key   text        NOT NULL, -- 2026-11 | 2026-11-26 | all
    used         integer     NOT NULL DEFAULT 0 CHECK (used >= 0),
    reserved     integer     NOT NULL DEFAULT 0 CHECK (reserved >= 0),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, counter_key, period_key)
);
CREATE INDEX idx_usage_counters_user ON usage_counters (user_id, period_key);

-- +goose Down
DROP TABLE IF EXISTS usage_counters;
