-- +goose Up
-- Global per-subject coaching records. Polymorphic by design: subject_type
-- names what the row is about and subject_id points at it, with no foreign
-- key. user_id is denormalized so "all my feedback" and "delete my data"
-- stay answerable. There is deliberately no `type` column: the
-- subject-to-product mapping is a product invariant (spec D18).
--
-- subject_id is uuid rather than text: the subject is always a row id and
-- every subject_type today is a uuid-keyed table. Text here made the
-- transcript read's `conversation_turns.id = feedbacks.subject_id` join
-- fail with "operator does not exist: uuid = text" (SQLSTATE 42883), which
-- GORM surfaces as an empty transcript rather than an error.
CREATE TABLE feedbacks (
    input_fingerprint text NOT NULL DEFAULT '',
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      text NOT NULL,
    subject_type text NOT NULL,
    subject_id   uuid NOT NULL,
    payload      jsonb NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_feedbacks_subject UNIQUE (subject_type, subject_id)
);

-- A future speech-progress report asks "which feedback has a speech block?".
-- GIN makes that a cheap containment query rather than a sequential scan.
CREATE INDEX idx_feedbacks_payload ON feedbacks USING GIN (payload);

CREATE INDEX idx_feedbacks_user_id ON feedbacks (user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS feedbacks;
