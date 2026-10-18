-- +goose Up
-- Repurpose the unused initial-schema user_noshow table as host-recorded no-shows per game.
-- Rows are only inserted ("Mark Absent") or deleted ("Mark Present"), so there is no updated_at.
ALTER TABLE user_noshow RENAME TO no_shows;
ALTER TABLE no_shows RENAME CONSTRAINT user_noshow_user_id_fkey TO no_shows_user_id_fkey;
ALTER TABLE no_shows RENAME CONSTRAINT user_noshow_event_id_fkey TO no_shows_event_id_fkey;

ALTER TABLE no_shows
  ADD COLUMN host_id UUID,
  ADD COLUMN marked_absent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE no_shows ns
SET host_id = eg.owner_id
FROM events e
JOIN event_groups eg ON eg.id = e.group_id
WHERE e.id = ns.event_id;

-- Drop rows whose event has no owning group; they could never be attributed to a host.
DELETE FROM no_shows WHERE host_id IS NULL;

ALTER TABLE no_shows ALTER COLUMN host_id SET NOT NULL;
ALTER TABLE no_shows ADD CONSTRAINT no_shows_host_id_fkey
  FOREIGN KEY (host_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE;

-- event_id leads so per-game absence lists are index-only scans on the PK.
ALTER TABLE no_shows DROP CONSTRAINT user_noshow_pkey;
ALTER TABLE no_shows ADD CONSTRAINT no_shows_pkey PRIMARY KEY (event_id, user_id);

-- Per-player, per-host absence counts (banner warning, future team-creation deprioritization)
-- and the users(id) cascade for user_id.
CREATE INDEX no_shows_user_id_host_id_idx ON no_shows (user_id, host_id);

-- +goose Down
DROP INDEX IF EXISTS no_shows_user_id_host_id_idx;
ALTER TABLE no_shows DROP CONSTRAINT no_shows_pkey;
ALTER TABLE no_shows ADD CONSTRAINT user_noshow_pkey PRIMARY KEY (user_id, event_id);
ALTER TABLE no_shows DROP CONSTRAINT no_shows_host_id_fkey;
ALTER TABLE no_shows
  DROP COLUMN created_at,
  DROP COLUMN marked_absent_at,
  DROP COLUMN host_id;
ALTER TABLE no_shows RENAME CONSTRAINT no_shows_event_id_fkey TO user_noshow_event_id_fkey;
ALTER TABLE no_shows RENAME CONSTRAINT no_shows_user_id_fkey TO user_noshow_user_id_fkey;
ALTER TABLE no_shows RENAME TO user_noshow;
