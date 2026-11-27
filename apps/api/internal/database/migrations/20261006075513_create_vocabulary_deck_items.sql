-- +goose Up
CREATE TABLE vocabulary_deck_items (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    deck_id             uuid NOT NULL REFERENCES vocabulary_decks (id) ON DELETE CASCADE,
    video_exercise_id   uuid REFERENCES video_exercises (id) ON DELETE SET NULL,
    video_transcript_id uuid REFERENCES video_transcripts (id) ON DELETE SET NULL,
    phrase              text NOT NULL,
    normalized_phrase   text NOT NULL,
    meaning             text NOT NULL,
    example_sentence    text,
    note                text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_vocabulary_deck_items_deck_id ON vocabulary_deck_items (deck_id);
CREATE INDEX idx_vocabulary_deck_items_video_exercise_id ON vocabulary_deck_items (video_exercise_id);
CREATE INDEX idx_vocabulary_deck_items_video_transcript_id ON vocabulary_deck_items (video_transcript_id);
CREATE INDEX idx_vocabulary_deck_items_normalized_phrase ON vocabulary_deck_items (normalized_phrase);

-- +goose Down
DROP TABLE IF EXISTS vocabulary_deck_items;
