-- name: GetRegistrationDataByEventId :many
SELECT R.*, 
    U.discord_name, 
    CASE 
		WHEN U.show_pronouns = true THEN pronouns 
		ELSE ''
	END AS pronouns, 
    GR.name AS current_rank_name 
FROM registrations AS R 
JOIN users AS U ON R.user_id = U.id 
JOIN events AS E ON R.event_id = E.id
JOIN event_groups AS EG ON E.group_id = EG.id
JOIN user_games AS UG ON R.user_id = UG.user_id AND EG.game_mode_id = UG.game_id
JOIN game_ranks AS GR ON UG.current_rank = GR.id
WHERE event_id = $1
ORDER BY U.discord_name ASC;