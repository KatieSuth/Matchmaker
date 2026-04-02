-- +goose Up
ALTER TABLE refresh_tokens DROP COLUMN revoked_at;

-- +goose Down
ALTER TABLE refresh_tokens ADD COLUMN revoked_at TIMESTAMP;