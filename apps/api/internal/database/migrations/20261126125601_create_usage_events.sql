-- +goose Up
-- Audit ledger + reservation record (spec §6.3). Rows are never deleted;
-- the only mutation is the status transition reserved -> committed |
-- released | expired. One row per counter charge: a voice session writes
-- two rows in one transaction. UNIQUE (user_id, idempotency_key) is the
-- double-charge guard.
CREATE TABLE usage_events (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    reservation_id   uuid        NOT NULL,
    user_id          text        NOT NULL,
    counter_key      text        NOT NULL REFERENCES premium_features(key),
    period_key       text        NOT NULL,
    action           text        NOT NULL
                                 CHECK (action IN ('voice_session_start','exam_ai_review',
                                                   'study_plan','writing_score','pronunciation')),
    amount           integer     NOT NULL CHECK (amount > 0),
    status           text        NOT NULL CHECK (status IN ('reserved','committed','released','expired')),
    idempotency_key  text        NOT NULL,
    ref_type         text        NOT NULL CHECK (ref_type IN ('conversation','exam_attempt','writing_activity','study_plan')),
    ref_id           text        NOT NULL DEFAULT '',
    expires_at       timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    settled_at       timestamptz,
    UNIQUE (user_id, idempotency_key)
);
CREATE INDEX idx_usage_events_expiry ON usage_events (status, expires_at);
CREATE INDEX idx_usage_events_reservation ON usage_events (reservation_id);
CREATE INDEX idx_usage_events_user ON usage_events (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS usage_events;
