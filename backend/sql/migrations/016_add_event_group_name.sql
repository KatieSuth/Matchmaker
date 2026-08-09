-- +goose Up
ALTER TABLE event_groups ADD COLUMN name TEXT
  CHECK (name IS NULL OR char_length(name) <= 50);

-- +goose Down
ALTER TABLE event_groups DROP COLUMN name;
