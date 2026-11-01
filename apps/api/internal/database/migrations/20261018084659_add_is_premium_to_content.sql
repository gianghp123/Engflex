-- +goose Up
-- Per-item premium gating for admin-curated content. Personal content
-- (user-owned vocabulary decks) never sets this — the column defaults to
-- false and nothing exposes it on the personal create/update paths.
ALTER TABLE lessons ADD COLUMN is_premium boolean NOT NULL DEFAULT false;
ALTER TABLE scenarios ADD COLUMN is_premium boolean NOT NULL DEFAULT false;
ALTER TABLE video_exercises ADD COLUMN is_premium boolean NOT NULL DEFAULT false;
ALTER TABLE vocabulary_items ADD COLUMN is_premium boolean NOT NULL DEFAULT false;
ALTER TABLE vocabulary_decks ADD COLUMN is_premium boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE vocabulary_decks DROP COLUMN is_premium;
ALTER TABLE vocabulary_items DROP COLUMN is_premium;
ALTER TABLE video_exercises DROP COLUMN is_premium;
ALTER TABLE scenarios DROP COLUMN is_premium;
ALTER TABLE lessons DROP COLUMN is_premium;
