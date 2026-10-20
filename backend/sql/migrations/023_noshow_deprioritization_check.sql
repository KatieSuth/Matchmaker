-- +goose Up
-- max_noshows is the no-show count (within 6 months, at this host's games) that triggers
-- deprioritization in team creation. It is only meaningful while deprioritize_noshows is on.
ALTER TABLE event_groups
  ADD CONSTRAINT event_groups_max_noshows_nonnegative CHECK (max_noshows >= 0),
  ADD CONSTRAINT event_groups_deprioritize_noshows_threshold CHECK (NOT deprioritize_noshows OR max_noshows >= 1);

-- +goose Down
ALTER TABLE event_groups
  DROP CONSTRAINT IF EXISTS event_groups_deprioritize_noshows_threshold,
  DROP CONSTRAINT IF EXISTS event_groups_max_noshows_nonnegative;
