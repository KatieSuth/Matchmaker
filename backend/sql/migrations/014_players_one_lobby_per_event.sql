-- +goose Up
ALTER TABLE players ADD COLUMN event_id UUID;

UPDATE players p
SET event_id = l.event_id
FROM lobbies l
WHERE l.id = p.lobby_id;

ALTER TABLE players ALTER COLUMN event_id SET NOT NULL;

ALTER TABLE players ADD CONSTRAINT players_event_id_fkey
  FOREIGN KEY (event_id) REFERENCES events(id)
  ON DELETE CASCADE ON UPDATE CASCADE;

CREATE UNIQUE INDEX players_event_id_user_id_uidx ON players (event_id, user_id);

-- +goose Down
DROP INDEX IF EXISTS players_event_id_user_id_uidx;
ALTER TABLE players DROP CONSTRAINT IF EXISTS players_event_id_fkey;
ALTER TABLE players DROP COLUMN event_id;
