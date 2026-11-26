-- +goose Up
-- One trial per Clerk user (spec §8.4): the user_id primary key is the
-- once-per-user guard; the application layer treats a row with ends_at in
-- the past as "trial used, free plan applies".
CREATE TABLE user_trials (
    user_id         text        PRIMARY KEY,
    started_at      timestamptz NOT NULL,
    ends_at         timestamptz NOT NULL,
    credits_granted integer     NOT NULL,
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS user_trials;
