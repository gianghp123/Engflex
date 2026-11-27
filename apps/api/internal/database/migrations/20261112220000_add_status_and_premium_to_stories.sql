-- +goose Up
ALTER TABLE stories ADD COLUMN status text NOT NULL DEFAULT 'draft'
    CONSTRAINT chk_stories_status CHECK (status IN ('draft', 'published'));
ALTER TABLE stories ADD COLUMN is_premium boolean NOT NULL DEFAULT false;
-- Preserve current visibility: everything seeded before this migration was
-- already live for learners, so backfill it published. Only rows an admin
-- creates from here on start as draft.
UPDATE stories SET status = 'published';

-- +goose Down
ALTER TABLE stories DROP COLUMN is_premium;
ALTER TABLE stories DROP COLUMN status;
