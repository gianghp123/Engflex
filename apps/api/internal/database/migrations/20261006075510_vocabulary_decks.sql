-- +goose Up
CREATE TABLE vocabulary_decks (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       text,
    category_id   uuid REFERENCES vocabulary_categories (id) ON DELETE SET NULL,
    name          text NOT NULL,
    description   text,
    thumbnail_url text,
    level         text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_vocabulary_decks_user_id ON vocabulary_decks (user_id);
CREATE INDEX idx_vocabulary_decks_category_id ON vocabulary_decks (category_id);
-- One name per user, case-insensitive. Curated decks (user_id NULL) are not
-- constrained to each other — NULLs never collide in a unique index anyway,
-- and the partial clause keeps the index to exactly the rows this rule owns.
-- Service-side TrimSpace keeps " daily words " colliding with "Daily Words".
CREATE UNIQUE INDEX uq_vocabulary_decks_user_name ON vocabulary_decks (user_id, lower(name)) WHERE user_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS vocabulary_decks;
