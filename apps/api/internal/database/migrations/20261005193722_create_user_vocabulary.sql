-- +goose Up
CREATE TABLE user_vocabulary (
    user_id          text NOT NULL,
    item_id          uuid NOT NULL REFERENCES vocabulary_items (id) ON DELETE CASCADE,
    source_type      text NOT NULL CONSTRAINT chk_user_vocabulary_source_type
        CHECK (source_type IN ('lesson', 'conversation', 'video', 'manual')),
    source_id        uuid,
    note             text,
    mastered         boolean NOT NULL DEFAULT false,
    srs_due_at       timestamptz,
    interval_days    integer NOT NULL DEFAULT 0,
    review_count     integer NOT NULL DEFAULT 0,
    last_reviewed_at timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, item_id)
);
CREATE INDEX idx_user_vocabulary_due ON user_vocabulary (user_id, srs_due_at);

-- +goose Down
DROP TABLE IF EXISTS user_vocabulary;
