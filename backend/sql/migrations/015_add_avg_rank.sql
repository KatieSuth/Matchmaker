-- +goose Up
ALTER TABLE user_games ADD COLUMN avg_rank UUID REFERENCES game_ranks(id);

UPDATE user_games ug
SET avg_rank = ar.id
FROM game_ranks cr, game_ranks pr, game_ranks ar
WHERE cr.id = ug.current_rank
  AND pr.id = ug.peak_rank
  AND ar.game_id = ug.game_id
  AND ar."order" = FLOOR((cr."order" + pr."order") / 2.0)::INT;

-- +goose Down
ALTER TABLE user_games DROP COLUMN avg_rank;
