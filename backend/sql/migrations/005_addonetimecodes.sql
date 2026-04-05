-- +goose Up
CREATE TABLE "one_time_codes" (
  "code" TEXT PRIMARY KEY,
  "user_id" UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  "expires" TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '60 seconds'
);

-- +goose Down
DROP TABLE one_time_codes