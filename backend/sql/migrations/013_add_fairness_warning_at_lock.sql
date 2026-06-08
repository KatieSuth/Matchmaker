-- +goose Up
ALTER TABLE lobbies ADD COLUMN fairness_warning_at_lock BOOLEAN NOT NULL DEFAULT false;
UPDATE lobbies SET fairness_warning_at_lock = fairness_warning;

-- +goose Down
ALTER TABLE lobbies DROP COLUMN fairness_warning_at_lock;
