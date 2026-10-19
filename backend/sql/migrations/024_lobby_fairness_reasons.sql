-- +goose Up
-- Which fairness check(s) the live fairness_warning tripped, so the UI can explain it.
-- Lobbies locked in before this migration keep both false; the UI falls back to generic copy.
ALTER TABLE lobbies
  ADD COLUMN fairness_outlier BOOLEAN NOT NULL DEFAULT false,
  ADD COLUMN fairness_team_separation BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE lobbies
  DROP COLUMN IF EXISTS fairness_team_separation,
  DROP COLUMN IF EXISTS fairness_outlier;
