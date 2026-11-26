-- +goose Up
-- Cached paid AI output (spec §6.4). The unique (subject_type, subject_id,
-- kind) is the caching boundary: generate once, read free forever.
CREATE TABLE ai_outputs (
    id               uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          text        NOT NULL,
    subject_type     text        NOT NULL CHECK (subject_type IN ('exam_attempt','writing_activity','study_plan')),
    subject_id       text        NOT NULL,
    kind             text        NOT NULL CHECK (kind IN ('exam_review','writing_review','study_plan','turn_analysis')),
    schema_version   integer     NOT NULL DEFAULT 1,
    payload          jsonb       NOT NULL,
    credits_charged  integer     NOT NULL DEFAULT 0,
    created_at       timestamptz NOT NULL DEFAULT now(),
    UNIQUE (subject_type, subject_id, kind)
);

-- +goose Down
DROP TABLE IF EXISTS ai_outputs;
