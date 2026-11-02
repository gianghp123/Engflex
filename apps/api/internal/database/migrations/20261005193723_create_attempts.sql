-- +goose Up
CREATE TABLE attempts (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    revision            integer NOT NULL DEFAULT 1 CHECK (revision >= 1),
    user_id             text NOT NULL,
    lesson_id           uuid REFERENCES lessons (id) ON DELETE SET NULL,
    video_exercise_id   uuid REFERENCES video_exercises (id) ON DELETE CASCADE,
    exam_test_id        uuid REFERENCES exam_tests (id) ON DELETE RESTRICT,
    type                text NOT NULL CONSTRAINT chk_attempts_type
        CHECK (type IN ('reading', 'listening', 'writing', 'speaking',
                        'shadowing', 'flashcard', 'benchmark', 'dictation', 'lesson', 'exam')),
    status              text NOT NULL DEFAULT 'in_progress' CONSTRAINT chk_attempts_status
        CHECK (status IN ('in_progress', 'completed', 'abandoned')),
    mode                text CONSTRAINT chk_attempts_mode
        CHECK (mode IS NULL OR mode IN ('exam', 'practice')),
    selected_parts      smallint[],
    score               numeric(5, 2),
    result              jsonb NOT NULL DEFAULT '{}',
    duration_sec        integer,
    started_at          timestamptz,
    submitted_at        timestamptz,
    deadline_at         timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT chk_attempts_exam_target
        CHECK (
            (type = 'exam' AND exam_test_id IS NOT NULL AND mode IS NOT NULL AND selected_parts IS NOT NULL AND started_at IS NOT NULL AND lesson_id IS NULL AND video_exercise_id IS NULL)
            OR
            (type <> 'exam' AND exam_test_id IS NULL AND mode IS NULL AND selected_parts IS NULL)
        ),
    CONSTRAINT chk_attempts_selected_parts
        CHECK (
            selected_parts IS NULL OR (
                array_ndims(selected_parts) = 1
                AND array_lower(selected_parts, 1) = 1
                AND cardinality(selected_parts) BETWEEN 1 AND 7
                AND array_position(selected_parts, NULL) IS NULL
                AND selected_parts <@ ARRAY[1,2,3,4,5,6,7]::smallint[]
                AND (
                    mode <> 'exam' OR (
                        cardinality(selected_parts) = 7
                        AND selected_parts @> ARRAY[1,2,3,4,5,6,7]::smallint[]
                    )
                )
            )
        ),
    CONSTRAINT chk_attempts_timeline
        CHECK (
            (type <> 'exam' OR status <> 'completed' OR submitted_at IS NOT NULL)
            AND (submitted_at IS NULL OR started_at IS NULL OR submitted_at >= started_at)
            AND (deadline_at IS NULL OR started_at IS NULL OR deadline_at >= started_at)
            AND (duration_sec IS NULL OR duration_sec >= 0)
        )
);
CREATE INDEX idx_attempts_user_created ON attempts (user_id, created_at DESC);
CREATE INDEX idx_attempts_user_type ON attempts (user_id, type);
CREATE INDEX idx_attempts_lesson_id ON attempts (lesson_id);
CREATE INDEX idx_attempts_video_exercise_id ON attempts (video_exercise_id);
CREATE INDEX idx_attempts_user_video_type_status ON attempts (user_id, video_exercise_id, type, status);
CREATE INDEX idx_attempts_exam_test_id ON attempts (exam_test_id);

CREATE UNIQUE INDEX uq_attempts_open_lesson ON attempts (user_id, lesson_id)
    WHERE type = 'lesson' AND status = 'in_progress';
CREATE UNIQUE INDEX uq_attempts_open_video ON attempts (user_id, video_exercise_id, type)
    WHERE video_exercise_id IS NOT NULL AND status = 'in_progress';

-- +goose Down
DROP TABLE IF EXISTS attempts;
