-- +goose Up
CREATE TABLE scenarios (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id   uuid NOT NULL REFERENCES scenario_topics (id) ON DELETE RESTRICT,
    persona_id uuid REFERENCES personas (id) ON DELETE SET NULL,
    title      text NOT NULL,
    objective  text NOT NULL DEFAULT '',
    cefr_level text NOT NULL CONSTRAINT chk_scenarios_cefr_level
        CHECK (cefr_level IN ('A1', 'A2', 'B1', 'B1+', 'B2', 'C1')),
    max_duration int NOT NULL,
    details    jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_scenarios_topic_id ON scenarios (topic_id);
CREATE INDEX idx_scenarios_persona_id ON scenarios (persona_id);

-- +goose Down
DROP TABLE IF EXISTS scenarios;
