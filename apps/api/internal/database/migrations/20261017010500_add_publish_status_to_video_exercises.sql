-- +goose Up
ALTER TABLE video_exercises ADD COLUMN status text NOT NULL DEFAULT 'draft'
    CONSTRAINT chk_video_exercises_status CHECK (status IN ('draft', 'published'));
-- Preserve current visibility: everything seeded before this migration was
-- already live for learners, so backfill it published. Only rows an admin
-- creates from here on start as draft.
UPDATE video_exercises SET status = 'published';

-- +goose Down
ALTER TABLE video_exercises DROP COLUMN status;
