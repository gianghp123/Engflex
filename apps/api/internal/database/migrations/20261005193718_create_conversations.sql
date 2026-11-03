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
    duration_sec          int,
    started_at            timestamptz NOT NULL DEFAULT now(),
    ended_at              timestamptz,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_conversations_user_started ON conversations (user_id, started_at DESC);
CREATE INDEX idx_conversations_scenario_id ON conversations (scenario_id);
CREATE INDEX idx_conversations_speech_session_id ON conversations (speech_session_id);

ALTER TABLE conversations ADD COLUMN IF NOT EXISTS objectives jsonb NOT NULL DEFAULT '[]';
-- snapshot of scenario.objectives at Create time (the scenario may change later)
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS summary jsonb;
-- {takeaways:string[], objectives:[{id,status:done|partial|missed,evidence}], levelNote}

-- +goose Down
DROP TABLE IF EXISTS conversations;
