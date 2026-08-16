-- +goose Up
ALTER TABLE lobbies ADD COLUMN join_code TEXT;
ALTER TABLE games ADD COLUMN join_link_base TEXT;

UPDATE games
SET join_link_base = 'https://gg.riotgames.com', updated_at = NOW()
WHERE name IN ('League of Legends', 'Valorant');

-- +goose Down
ALTER TABLE lobbies DROP COLUMN join_code;
ALTER TABLE games DROP COLUMN join_link_base;
