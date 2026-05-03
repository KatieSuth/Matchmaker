-- +goose Up
ALTER TABLE events ADD COLUMN game_mode_id UUID REFERENCES game_modes(id) ON DELETE RESTRICT ON UPDATE CASCADE;

UPDATE events SET game_mode_id = (
    SELECT eg.game_mode_id FROM event_groups eg WHERE eg.id = events.group_id
);

ALTER TABLE events ALTER COLUMN game_mode_id SET NOT NULL;

ALTER TABLE event_groups DROP CONSTRAINT event_groups_game_mode_id_fkey;
ALTER TABLE event_groups DROP COLUMN game_mode_id;

-- +goose Down
ALTER TABLE event_groups ADD COLUMN game_mode_id UUID;

UPDATE event_groups eg SET game_mode_id = (
    SELECT e.game_mode_id FROM events e WHERE e.group_id = eg.id ORDER BY e.start_time ASC NULLS LAST LIMIT 1
);

ALTER TABLE event_groups ALTER COLUMN game_mode_id SET NOT NULL;
ALTER TABLE event_groups ADD CONSTRAINT event_groups_game_mode_id_fkey FOREIGN KEY (game_mode_id) REFERENCES game_modes(id) ON DELETE SET NULL ON UPDATE CASCADE;

ALTER TABLE events DROP CONSTRAINT events_game_mode_id_fkey;
ALTER TABLE events DROP COLUMN game_mode_id;
