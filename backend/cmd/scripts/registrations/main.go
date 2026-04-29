// Command registrations seeds development registrations for local testing only.
// Run order: run after users and events seeders.
package main

import (
	"log/slog"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/google/uuid"
)

const (
	minRegistrations = 9
	maxRegistrations = 26
)

func main() {
	seed := common.NewSeedContext()
	defer seed.Close()

	var seededUsers []uuid.UUID
	userRows, err := seed.Pool.Query(seed.Ctx, `
		SELECT id
		FROM users
		WHERE discord_id LIKE 'seed_user_%'
		ORDER BY discord_id ASC
	`)
	if err != nil {
		common.Fatal("failed querying seeded users", "error", err)
	}
	for userRows.Next() {
		var id uuid.UUID
		if err := userRows.Scan(&id); err != nil {
			userRows.Close()
			common.Fatal("failed scanning seeded user", "error", err)
		}
		seededUsers = append(seededUsers, id)
	}
	if err := userRows.Err(); err != nil {
		userRows.Close()
		common.Fatal("failed iterating seeded users", "error", err)
	}
	userRows.Close()

	if len(seededUsers) < 30 {
		common.Fatal("missing prerequisite seeded users; run users script first", "required", 30, "actual", len(seededUsers))
	}
	if len(seededUsers) < maxRegistrations {
		common.Fatal("insufficient unique users for max registration target", "required", maxRegistrations, "actual", len(seededUsers))
	}

	var events []uuid.UUID
	eventRows, err := seed.Pool.Query(seed.Ctx, `
		SELECT e.id
		FROM events e
		JOIN event_groups eg ON eg.id = e.group_id
		JOIN users u ON u.id = eg.owner_id
		WHERE u.discord_id LIKE 'seed_user_%'
		ORDER BY e.start_time ASC, e.id ASC
	`)
	if err != nil {
		common.Fatal("failed querying seeded events", "error", err)
	}
	for eventRows.Next() {
		var id uuid.UUID
		if err := eventRows.Scan(&id); err != nil {
			eventRows.Close()
			common.Fatal("failed scanning seeded event", "error", err)
		}
		events = append(events, id)
	}
	if err := eventRows.Err(); err != nil {
		eventRows.Close()
		common.Fatal("failed iterating seeded events", "error", err)
	}
	eventRows.Close()

	if len(events) == 0 {
		common.Fatal("missing prerequisite seeded events; run events script first")
	}

	created := 0
	for i, eventID := range events {
		target := minRegistrations + ((i * 7) % (maxRegistrations - minRegistrations + 1))
		offset := (i * 5) % len(seededUsers)

		for n := 0; n < target; n++ {
			userID := seededUsers[(offset+n)%len(seededUsers)]
			canSubstitute := (n%3 == 0)
			canLobbyHost := (n == 0) || (n%11 == 0)

			var duoRequest *string
			if n%8 == 0 {
				name := "SeedUser001"
				duoRequest = &name
			}

			tag, err := seed.Pool.Exec(seed.Ctx, `
				INSERT INTO registrations (event_id, user_id, can_substitute, can_lobby_host, duo_request, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
				ON CONFLICT (event_id, user_id) DO NOTHING
			`, eventID, userID, canSubstitute, canLobbyHost, duoRequest)
			if err != nil {
				common.Fatal("failed creating registration", "event_id", eventID, "user_id", userID, "error", err)
			}
			created += int(tag.RowsAffected())
		}
	}

	rows, err := seed.Pool.Query(seed.Ctx, `
		SELECT e.id, COUNT(r.user_id)::INT
		FROM events e
		JOIN event_groups eg ON eg.id = e.group_id
		JOIN users u ON u.id = eg.owner_id
		LEFT JOIN registrations r ON r.event_id = e.id
		WHERE u.discord_id LIKE 'seed_user_%'
		GROUP BY e.id
	`)
	if err != nil {
		common.Fatal("failed validating registration counts", "error", err)
	}
	for rows.Next() {
		var eventID uuid.UUID
		var count int
		if err := rows.Scan(&eventID, &count); err != nil {
			rows.Close()
			common.Fatal("failed scanning registration validation row", "error", err)
		}
		if count < minRegistrations || count > maxRegistrations {
			rows.Close()
			common.Fatal("registration count out of requested range", "event_id", eventID, "min", minRegistrations, "max", maxRegistrations, "actual", count)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		common.Fatal("failed iterating registration validation rows", "error", err)
	}
	rows.Close()

	slog.Info("registration seeding complete", "events", len(events), "new_rows_created", created, "count_range", "9-26")
}
