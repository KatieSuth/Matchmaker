-- +goose Up
ALTER TABLE users ADD COLUMN display_name TEXT
  CHECK (display_name IS NULL OR char_length(display_name) <= 50);

-- +goose Down
ALTER TABLE users DROP COLUMN display_name;
