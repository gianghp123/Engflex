-- +goose Up
CREATE TABLE bookmarks (
    user_id      text NOT NULL,
    content_type text NOT NULL CONSTRAINT chk_bookmarks_content_type
        CHECK (content_type IN ('lesson', 'video', 'story')),
    content_id   uuid NOT NULL,
    note         text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, content_type, content_id)
);
CREATE INDEX idx_bookmarks_user ON bookmarks (user_id, content_type, created_at DESC);
CREATE INDEX idx_bookmarks_content ON bookmarks (content_type, content_id);

-- +goose Down
DROP TABLE IF EXISTS bookmarks;
