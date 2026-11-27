-- +goose Up
-- Admin-editable AI-credit price per paid action. Only the three paid actions
-- are representable: writing_score and pronunciation stay free in code.
CREATE TABLE action_credit_costs (
    action      text        PRIMARY KEY
                            CHECK (action IN ('voice_session_start','exam_ai_review','study_plan')),
    credits     integer     NOT NULL CHECK (credits BETWEEN 0 AND 100000),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    updated_by  text        NOT NULL DEFAULT 'system'
);

INSERT INTO action_credit_costs (action, credits) VALUES
    ('voice_session_start', 1),
    ('exam_ai_review', 1),
    ('study_plan', 2);

-- +goose Down
DROP TABLE IF EXISTS action_credit_costs;
