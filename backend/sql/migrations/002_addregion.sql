-- +goose Up
ALTER TABLE users ADD COLUMN region TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN region;