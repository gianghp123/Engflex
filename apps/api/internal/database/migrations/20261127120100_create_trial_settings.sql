-- +goose Up
-- Trial duration is admin-editable; credit/voice caps stay in
-- plan_feature_limits (the enforcement source). Single row, id enforced true.
CREATE TABLE trial_settings (
    id            boolean     PRIMARY KEY DEFAULT true CHECK (id),
    duration_days integer     NOT NULL CHECK (duration_days BETWEEN 1 AND 365),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    updated_by    text        NOT NULL DEFAULT 'system'
);

INSERT INTO trial_settings (id, duration_days) VALUES (true, 7);

-- +goose Down
DROP TABLE IF EXISTS trial_settings;
