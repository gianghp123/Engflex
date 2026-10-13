-- +goose Up
ALTER TABLE vocabulary_items ADD COLUMN status text NOT NULL DEFAULT 'draft'
    CONSTRAINT chk_vocabulary_items_status CHECK (status IN ('draft', 'published'));
-- Preserve current visibility: everything seeded before this migration was
-- already live for learners, so backfill it published. Only rows an admin
-- creates from here on start as draft.
UPDATE vocabulary_items SET status = 'published';

ALTER TABLE vocabulary_decks ADD COLUMN status text NOT NULL DEFAULT 'draft'
    CONSTRAINT chk_vocabulary_decks_status CHECK (status IN ('draft', 'published'));
UPDATE vocabulary_decks SET status = 'published';

-- +goose Down
ALTER TABLE vocabulary_decks DROP COLUMN status;
ALTER TABLE vocabulary_items DROP COLUMN status;
