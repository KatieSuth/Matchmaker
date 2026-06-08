-- +goose Up
ALTER TABLE lobbies ADD COLUMN fairness_warning BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE lobbies DROP COLUMN fairness_warning;
