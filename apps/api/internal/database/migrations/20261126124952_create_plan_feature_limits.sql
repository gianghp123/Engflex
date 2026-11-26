-- +goose Up
-- Entitlement catalog: per-plan counter limits (spec §6.1). plan_feature_limits
-- stores ONLY counter limits; premium_features stays the feature registry
-- (label/description/admin toggle) and is seeded with the two counters here
-- because its key is this table's foreign key.

INSERT INTO premium_features (key, label, description) VALUES
    ('ai_credit', 'AI credits', 'One unit of metered AI usage: a voice session, an exam review, or a study plan.'),
    ('voice_session', 'Voice sessions', 'Cost sub-cap: one started voice conversation, up to the engine duration cap.')
ON CONFLICT (key) DO NOTHING;

CREATE TABLE plan_feature_limits (
    plan_key     text        NOT NULL CHECK (plan_key IN ('free','trial','1m','3m','12m','admin')),
    counter_key  text        NOT NULL REFERENCES premium_features(key),
    is_enabled   boolean     NOT NULL DEFAULT true,
    limit_value  integer     NOT NULL DEFAULT 0 CHECK (limit_value >= 0),
    period       text        NOT NULL DEFAULT 'month' CHECK (period IN ('day','month','total')),
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (plan_key, counter_key)
);

-- Launch values (spec §5.5). Trial limits are one-time totals; admin rows
-- are documentation-only (the policy short-circuits admins before limits).
INSERT INTO plan_feature_limits (plan_key, counter_key, is_enabled, limit_value, period) VALUES
    ('free',  'ai_credit',     true, 15, 'month'),
    ('free',  'voice_session', true,  3, 'month'),
    ('trial', 'ai_credit',     true, 30, 'total'),
    ('trial', 'voice_session', true,  7, 'total'),
    ('1m',    'ai_credit',     true, 100, 'month'),
    ('1m',    'voice_session', true,  60, 'month'),
    ('3m',    'ai_credit',     true, 100, 'month'),
    ('3m',    'voice_session', true,  60, 'month'),
    ('12m',   'ai_credit',     true, 100, 'month'),
    ('12m',   'voice_session', true,  60, 'month'),
    ('admin', 'ai_credit',     false, 0, 'total'),
    ('admin', 'voice_session', false, 0, 'total');

-- +goose Down
DROP TABLE IF EXISTS plan_feature_limits;
DELETE FROM premium_features WHERE key IN ('ai_credit', 'voice_session');
