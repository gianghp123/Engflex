-- +goose Up
CREATE TABLE vocabulary_items (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    term               text NOT NULL,
    -- term_key is the word's identity: trimmed, inner whitespace collapsed,
    -- case-folded. Generated, so no writer can forge it and no row can carry a
    -- NULL key (a NULL would escape the unique indexes below).
    term_key           text GENERATED ALWAYS AS
        (lower(regexp_replace(btrim(term), '\s+', ' ', 'g'))) STORED,
    definition         text NOT NULL DEFAULT '',
    cefr               text NOT NULL CONSTRAINT chk_vocabulary_items_cefr
        CHECK (cefr IN ('A1', 'A2', 'B1', 'B2', 'C1', 'C2')),
    part_of_speech     text CONSTRAINT chk_vocabulary_items_part_of_speech
        CHECK (part_of_speech IS NULL OR part_of_speech IN
            ('noun', 'verb', 'adjective', 'adverb', 'pronoun', 'preposition',
             'conjunction', 'interjection', 'determiner', 'phrase', 'idiom',
             'phrasal_verb')),
    created_by_user_id text,
    details            jsonb NOT NULL DEFAULT '{}',
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uq_vocabulary_items_global_term
    ON vocabulary_items (term_key) WHERE created_by_user_id IS NULL;
CREATE UNIQUE INDEX uq_vocabulary_items_user_term
    ON vocabulary_items (created_by_user_id, term_key) WHERE created_by_user_id IS NOT NULL;
CREATE INDEX idx_vocabulary_items_cefr ON vocabulary_items (cefr);

-- +goose Down
DROP TABLE IF EXISTS vocabulary_items;
