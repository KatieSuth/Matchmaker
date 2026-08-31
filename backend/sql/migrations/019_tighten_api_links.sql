-- +goose Up
-- api_links is unused; tighten it for encrypted per-provider refresh tokens.
DROP INDEX IF EXISTS api_links_user_id_name_idx;

ALTER TABLE api_links
  ALTER COLUMN user_id SET NOT NULL,
  ALTER COLUMN name SET NOT NULL,
  ALTER COLUMN refresh_token SET NOT NULL,
  ALTER COLUMN refresh_token_iv SET NOT NULL;

CREATE UNIQUE INDEX api_links_user_id_name_key ON api_links (user_id, name);

-- +goose Down
DROP INDEX IF EXISTS api_links_user_id_name_key;

ALTER TABLE api_links
  ALTER COLUMN user_id DROP NOT NULL,
  ALTER COLUMN name DROP NOT NULL,
  ALTER COLUMN refresh_token DROP NOT NULL,
  ALTER COLUMN refresh_token_iv DROP NOT NULL;

CREATE INDEX api_links_user_id_name_idx ON api_links (user_id, name);
