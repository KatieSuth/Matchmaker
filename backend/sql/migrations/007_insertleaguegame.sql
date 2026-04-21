-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    v_game_id UUID;
BEGIN
    INSERT INTO games (id, "name", created_at, updated_at)
    VALUES (gen_random_uuid(), 'League of Legends', NOW(), NOW())
    RETURNING id INTO v_game_id;

    INSERT INTO game_modes (id, game_id, "name", team_size, created_at, updated_at, duration)
    VALUES (gen_random_uuid(), v_game_id, '5v5', 5, NOW(), NOW(), 60);

    INSERT INTO game_ranks (id, game_id, "name", "order", created_at, updated_at)
    VALUES
        (gen_random_uuid(), v_game_id, 'Iron',        1, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze',      2, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver',      3, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold',        4, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum',    5, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Emerald',     6, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond',     7, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Master',      8, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Grandmaster', 9, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Challenger', 10, NOW(), NOW());
END $$;
-- +goose StatementEnd


-- +goose Down
DELETE FROM games WHERE "name" = "League of Legends"

--(game_modes and game_ranks will be deleted if their game_id is delete)