-- +goose Up
-- +goose StatementBegin
-- Replace League of Legends ranks with tier divisions (IV–I) plus Master / Grandmaster / Challenger.
-- Existing user_games rank FKs clear via ON DELETE SET NULL when old rank rows are removed.
DO $$
DECLARE
    v_game_id UUID;
BEGIN
    SELECT id INTO v_game_id FROM games WHERE name = 'League of Legends' LIMIT 1;
    IF NOT FOUND THEN
        RETURN;
    END IF;

    DELETE FROM game_ranks WHERE game_id = v_game_id;

    INSERT INTO game_ranks (id, game_id, "name", "order", created_at, updated_at)
    VALUES
        (gen_random_uuid(), v_game_id, 'Iron IV',          1, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Iron III',         2, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Iron II',          3, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Iron I',           4, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze IV',        5, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze III',       6, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze II',        7, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Bronze I',         8, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver IV',        9, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver III',      10, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver II',       11, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Silver I',        12, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold IV',         13, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold III',        14, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold II',         15, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Gold I',          16, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum IV',     17, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum III',    18, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum II',     19, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Platinum I',      20, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Emerald IV',      21, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Emerald III',     22, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Emerald II',      23, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Emerald I',       24, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond IV',      25, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond III',     26, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond II',      27, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Diamond I',       28, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Master',          29, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Grandmaster',     30, NOW(), NOW()),
        (gen_random_uuid(), v_game_id, 'Challenger',      31, NOW(), NOW());
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
    v_game_id UUID;
BEGIN
    SELECT id INTO v_game_id FROM games WHERE name = 'League of Legends' LIMIT 1;
    IF NOT FOUND THEN
        RETURN;
    END IF;

    DELETE FROM game_ranks WHERE game_id = v_game_id;

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
