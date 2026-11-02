-- +goose Up
CREATE TABLE conversations (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               text NOT NULL,
    mode                  text NOT NULL CONSTRAINT chk_conversations_mode
        CHECK (mode IN ('free_talk', 'roleplay')),
    scenario_id           uuid REFERENCES scenarios (id) ON DELETE SET NULL,
    status                text NOT NULL CONSTRAINT chk_conversations_status
        CHECK (status IN ('pending', 'live', 'ended', 'failed')),
    speech_session_id     text,
    speech_start_response jsonb,
    transcript_revision bigint NOT NULL DEFAULT 0 CHECK (transcript_revision >= 0),
    transcript_hash text NOT NULL DEFAULT '',
    transcript_finalized boolean NOT NULL DEFAULT false,
    duration_sec          int CHECK (duration_sec >= 0),
    started_at            timestamptz NOT NULL DEFAULT now(),
    ended_at              timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_conversations_user_started ON conversations (user_id, started_at DESC);
CREATE INDEX idx_conversations_scenario_id ON conversations (scenario_id);
CREATE INDEX idx_conversations_speech_session_id ON conversations (speech_session_id);

CREATE UNIQUE INDEX uq_conversations_active_user ON conversations (user_id)
    WHERE status IN ('pending', 'live');

-- +goose Down
DROP TABLE IF EXISTS conversations;
