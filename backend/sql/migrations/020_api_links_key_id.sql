-- +goose Up
-- key_id selects which AES key in the process keyring sealed this row (rotation without a flag day).
ALTER TABLE api_links
  ADD COLUMN key_id text NOT NULL DEFAULT '1';

-- +goose Down
ALTER TABLE api_links
  DROP COLUMN key_id;
