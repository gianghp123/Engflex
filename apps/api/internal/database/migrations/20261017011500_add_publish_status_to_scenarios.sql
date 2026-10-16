-- +goose Up
ALTER TABLE scenarios ADD COLUMN status text NOT NULL DEFAULT 'draft'
    CONSTRAINT chk_scenarios_status CHECK (status IN ('draft', 'published'));
-- Preserve current visibility: everything seeded before this migration was
-- already live for learners, so backfill it published. Only rows an admin
-- creates from here on start as draft.
UPDATE scenarios SET status = 'published';

-- +goose Down
ALTER TABLE scenarios DROP COLUMN status;
