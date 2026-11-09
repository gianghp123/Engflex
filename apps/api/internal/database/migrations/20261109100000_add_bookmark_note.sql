-- +goose Up
ALTER TABLE bookmarks ADD COLUMN note text;

-- +goose Down
ALTER TABLE bookmarks DROP COLUMN note;
