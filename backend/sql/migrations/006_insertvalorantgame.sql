-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    v_game_id UUID;
BEGIN
    INSERT INTO games (id, "name", created_at, updated_at)
    VALUES (gen_random_uuid(), 'Valorant', NOW(), NOW())
    RETURNING id INTO v_game_id;

    INSERT INTO game_modes (id, game_id, "name", team_size, created_at, updated_at)
    VALUES (gen_random_uuid(), v_game_id, '5v5', 5, NOW(), NOW());

    INSERT INTO game_ranks (id, game_id, "name", "order", created_at, updated_at)
    VALUES
        (gen_random_uuid(), v_game_id, 'Iron 1',       1,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Iron 2',       2,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Iron 3',       3,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze 1',     4,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze 2',     5,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze 3',     6,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver 1',     7,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver 2',     8,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver 3',     9,  NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold 1',       10, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold 2',       11, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold 3',       12, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum 1',   13, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum 2',   14, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum 3',   15, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond 1',    16, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond 2',    17, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond 3',    18, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Ascendant 1',  19, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Ascendant 2',  20, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Ascendant 3',  21, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Immortal 1',   22, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Immortal 2',   23, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Immortal 3',   24, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Radiant',      25, NOW(), NOW());
END $$;
-- +goose StatementEnd


-- +goose Down
DELETE FROM games WHERE "name" = "Valorant"

--(game_modes and game_ranks will be deleted if their game_id is delete)