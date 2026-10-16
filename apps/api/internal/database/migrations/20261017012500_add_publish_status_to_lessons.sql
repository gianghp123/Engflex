-- +goose Up
ALTER TABLE lessons ADD COLUMN status text NOT NULL DEFAULT 'draft'
    CONSTRAINT chk_lessons_status CHECK (status IN ('draft', 'published'));
-- Preserve current visibility: everything seeded before this migration was
-- already live for learners, so backfill it published. Only rows an admin
-- creates from here on start as draft.
UPDATE lessons SET status = 'published';

-- +goose Down
ALTER TABLE lessons DROP COLUMN status;
