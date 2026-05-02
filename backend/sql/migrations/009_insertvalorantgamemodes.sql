-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    v_game_id UUID;
BEGIN
    SELECT id INTO v_game_id FROM games WHERE "name" = 'Valorant';

    INSERT INTO game_modes (id, game_id, "name", team_size, owner_id, created_at, updated_at, duration)
    VALUES
        (gen_random_uuid(), v_game_id, '4v4 on A/B Map', 4, NULL, NOW(), NOW(), 60),
        (gen_random_uuid(), v_game_id, '3v3 Skirmish', 3, NULL, NOW(), NOW(), 15),
        (gen_random_uuid(), v_game_id, '2v2 Skirmish', 2, NULL, NOW(), NOW(), 15);
END $$;
-- +goose StatementEnd


-- +goose Down
DELETE FROM game_modes
WHERE game_id = (SELECT id FROM games WHERE "name" = 'Valorant')
  AND "name" IN ('4v4 on A/B Map', '3v3 Skirmish', '2v2 Skirmish')
  AND owner_id IS NULL;
