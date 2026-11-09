-- +goose Up
-- Allow 'ielts' exam type in exam_categories
ALTER TABLE exam_categories DROP CONSTRAINT IF EXISTS chk_exam_categories_type;
ALTER TABLE exam_categories ADD CONSTRAINT chk_exam_categories_type
    CHECK (type IN ('toeic', 'ielts'));

-- Add sub_type (academic, general_training) to exam_categories
ALTER TABLE exam_categories ADD COLUMN sub_type text CONSTRAINT chk_exam_categories_sub_type
    CHECK (sub_type IS NULL OR sub_type IN ('academic', 'general_training'));

-- Add skill and sub_type to exam_tests with closed set check constraints
ALTER TABLE exam_tests ADD COLUMN skill text CONSTRAINT chk_exam_tests_skill
    CHECK (skill IS NULL OR skill IN ('reading', 'listening', 'writing', 'speaking', 'full'));

ALTER TABLE exam_tests ADD COLUMN sub_type text CONSTRAINT chk_exam_tests_sub_type
    CHECK (sub_type IS NULL OR sub_type IN ('academic', 'general_training'));

CREATE INDEX idx_exam_tests_skill_sub_type ON exam_tests (skill, sub_type);

-- Relax attempts.selected_parts constraint: retain array structure 1-7 validity,
-- while delegating exam-specific part count enforcement (7 parts for TOEIC, 3-4 sections for IELTS) to service layer.
ALTER TABLE attempts DROP CONSTRAINT IF EXISTS chk_attempts_selected_parts;
ALTER TABLE attempts ADD CONSTRAINT chk_attempts_selected_parts
    CHECK (
        selected_parts IS NULL OR (
            array_ndims(selected_parts) = 1
            AND array_lower(selected_parts, 1) = 1
            AND cardinality(selected_parts) BETWEEN 1 AND 7
            AND array_position(selected_parts, NULL) IS NULL
            AND selected_parts <@ ARRAY[1,2,3,4,5,6,7]::smallint[]
        )
    );

-- +goose Down
-- Revert attempts.selected_parts constraint back to strict TOEIC-only cardinality
ALTER TABLE attempts DROP CONSTRAINT IF EXISTS chk_attempts_selected_parts;
ALTER TABLE attempts ADD CONSTRAINT chk_attempts_selected_parts
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
    );

DROP INDEX IF EXISTS idx_exam_tests_skill_sub_type;
ALTER TABLE exam_tests DROP COLUMN IF EXISTS sub_type;
ALTER TABLE exam_tests DROP COLUMN IF EXISTS skill;

ALTER TABLE exam_categories DROP COLUMN IF EXISTS sub_type;
ALTER TABLE exam_categories DROP CONSTRAINT IF EXISTS chk_exam_categories_type;
ALTER TABLE exam_categories ADD CONSTRAINT chk_exam_categories_type
    CHECK (type = 'toeic');
