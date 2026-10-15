-- +goose Up
CREATE TABLE vocabulary_deck_items (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    deck_id          uuid NOT NULL REFERENCES vocabulary_decks (id) ON DELETE CASCADE,
    item_id          uuid NOT NULL REFERENCES vocabulary_items (id) ON DELETE CASCADE,
    sequence         integer NOT NULL DEFAULT 0,
    example_sentence text,
    note             text,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uq_vocabulary_deck_items_deck_item UNIQUE (deck_id, item_id)
);
CREATE INDEX idx_vocabulary_deck_items_deck_id ON vocabulary_deck_items (deck_id);
CREATE INDEX idx_vocabulary_deck_items_item_id ON vocabulary_deck_items (item_id);

-- +goose Down
DROP TABLE IF EXISTS vocabulary_deck_items;
