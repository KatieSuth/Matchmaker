-- +goose Up
ALTER TABLE users ADD COLUMN new_user BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE users DROP COLUMN new_user;