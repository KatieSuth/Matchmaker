// Command events seeds development event groups/events for local testing only.
// Run order: run after users seeder and before registrations seeder.
package main

import (
	"log/slog"
	"sort"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

const (
	totalGroups     = 20
	pastStartGroups = 3
)

var seedRegions = []string{"AMER", "EMEA", "APAC"}

// IANA zones used so offsets match region (AMER ≈ -04…-07, EMEA European, APAC Australian).
var amerTimezones = []string{
	"America/New_York",
	"America/Chicago",
	"America/Denver",
	"America/Los_Angeles",
}
var emeaTimezones = []string{"Europe/London", "Europe/Berlin", "Europe/Paris"}
var apacTimezones = []string{"Australia/Sydney", "Australia/Melbourne", "Australia/Brisbane"}

type gameModeSeed struct {
	ID       uuid.UUID
	Duration int32
}

func seedTimezoneForRegion(region string, groupIndex int) *time.Location {
	var name string
	switch region {
	case "AMER":
		name = amerTimezones[groupIndex%len(amerTimezones)]
	case "EMEA":
		name = emeaTimezones[groupIndex%len(emeaTimezones)]
	case "APAC":
		name = apacTimezones[groupIndex%len(apacTimezones)]
	default:
		common.Fatal("unknown seed region", "region", region)
		panic("unreachable")
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		common.Fatal("failed loading timezone", "name", name, "error", err)
	}
	return loc
}

// firstLocalSevenPMOnOrAfter returns the next calendar evening at 19:00 in loc at or after ref.
func firstLocalSevenPMOnOrAfter(loc *time.Location, ref time.Time) time.Time {
	refLocal := ref.In(loc)
	y, m, d := refLocal.Date()
	today7 := time.Date(y, m, d, 19, 0, 0, 0, loc)
	if !today7.Before(refLocal) {
		return today7
	}
	next := time.Date(y, m, d, 0, 0, 0, 0, loc).AddDate(0, 0, 1)
	return time.Date(next.Year(), next.Month(), next.Day(), 19, 0, 0, 0, loc)
}

// seedFirstEventStart picks the first event start for a group: 19:00 local on a recent past day,
// or the next and following evenings at 19:00 local for upcoming groups.
func seedFirstEventStart(loc *time.Location, groupIndex int, past bool) time.Time {
	if past {
		now := time.Now().In(loc)
		y, m, d := now.Date()
		today := time.Date(y, m, d, 0, 0, 0, 0, loc)
		day := today.AddDate(0, 0, -(groupIndex + 1))
		return time.Date(day.Year(), day.Month(), day.Day(), 19, 0, 0, 0, loc)
	}
	j := groupIndex - pastStartGroups
	base := firstLocalSevenPMOnOrAfter(loc, time.Now())
	return base.AddDate(0, 0, j)
}

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

	var gameModes []gameModeSeed
	modeRows, err := seed.Pool.Query(seed.Ctx, `
		WITH ranked_modes AS (
			SELECT
				gm.id,
				gm.duration,
				gm.game_id,
				ROW_NUMBER() OVER (PARTITION BY gm.game_id ORDER BY gm.name ASC, gm.id ASC) AS rn
			FROM game_modes gm
			JOIN games g ON g.id = gm.game_id
			WHERE g.owner_id IS NULL
		)
		SELECT id, duration
		FROM ranked_modes
		WHERE rn = 1
		ORDER BY game_id ASC
	`)
	if err != nil {
		common.Fatal("failed querying available game modes", "error", err)
	}
	for modeRows.Next() {
		var gm gameModeSeed
		if err := modeRows.Scan(&gm.ID, &gm.Duration); err != nil {
			modeRows.Close()
			common.Fatal("failed scanning game mode", "error", err)
		}
		gameModes = append(gameModes, gm)
	}
	if err := modeRows.Err(); err != nil {
		modeRows.Close()
		common.Fatal("failed iterating game modes", "error", err)
	}
	modeRows.Close()

	if len(gameModes) == 0 {
		common.Fatal("missing available games/game modes; run migrations/seeds for games first")
	}

	var existingGroupCount int
	if err := seed.Pool.QueryRow(seed.Ctx, `
		SELECT COUNT(*)
		FROM event_groups eg
		JOIN users u ON u.id = eg.owner_id
		WHERE u.discord_id LIKE 'seed_user_%'
	`).Scan(&existingGroupCount); err != nil {
		common.Fatal("failed checking existing seeded event groups", "error", err)
	}

	if existingGroupCount == totalGroups {
		slog.Info("event groups already seeded; skipping create", "groups", existingGroupCount)
		return
	}
	if existingGroupCount > 0 && existingGroupCount != totalGroups {
		common.Fatal("partial seeded event groups found; cleanup required before re-seeding", "found", existingGroupCount, "expected", totalGroups)
	}

	createdGroups := 0
	createdEvents := 0

	for i := 0; i < totalGroups; i++ {
		ownerID := seededUsers[i%len(seededUsers)]
		mode := gameModes[i%len(gameModes)]
		gamesToRun := int32((i % 3) + 1) // 1-3 events per group

		groupRegion := seedRegions[i%len(seedRegions)]
		loc := seedTimezoneForRegion(groupRegion, i)
		start := seedFirstEventStart(loc, i, i < pastStartGroups)
		registrationOpen := i >= pastStartGroups
		group, err := seed.Queries.CreateEventGroup(seed.Ctx, db.CreateEventGroupParams{
			OwnerID:          ownerID,
			GameModeID:       mode.ID,
			SubMin:           int32(i % 6),
			RegistrationOpen: registrationOpen,
			Region:           groupRegion,
		})
		if err != nil {
			common.Fatal("failed creating event group", "index", i, "error", err)
		}
		createdGroups++

		nextStart := start
		for n := int32(0); n < gamesToRun; n++ {
			groupID := group.ID
			if err := seed.Queries.CreateEvent(seed.Ctx, db.CreateEventParams{
				GroupID:   &groupID,
				StartTime: nextStart,
			}); err != nil {
				common.Fatal("failed creating event", "group_id", group.ID, "event_index", n, "error", err)
			}
			createdEvents++
			nextStart = nextStart.Add(time.Duration(mode.Duration) * time.Minute)
		}
	}

	type checkRow struct {
		GroupID   uuid.UUID
		Duration  int32
		StartTime time.Time
	}
	checkRows, err := seed.Pool.Query(seed.Ctx, `
		SELECT eg.id, gm.duration, e.start_time
		FROM event_groups eg
		JOIN users u ON u.id = eg.owner_id
		JOIN game_modes gm ON gm.id = eg.game_mode_id
		JOIN events e ON e.group_id = eg.id
		WHERE u.discord_id LIKE 'seed_user_%'
		ORDER BY eg.id, e.start_time
	`)
	if err != nil {
		common.Fatal("failed validating seeded event timing", "error", err)
	}

	byGroup := map[uuid.UUID][]checkRow{}
	for checkRows.Next() {
		var row checkRow
		if err := checkRows.Scan(&row.GroupID, &row.Duration, &row.StartTime); err != nil {
			checkRows.Close()
			common.Fatal("failed scanning seeded event validation row", "error", err)
		}
		byGroup[row.GroupID] = append(byGroup[row.GroupID], row)
	}
	if err := checkRows.Err(); err != nil {
		checkRows.Close()
		common.Fatal("failed iterating seeded event validation rows", "error", err)
	}
	checkRows.Close()

	if len(byGroup) != totalGroups {
		common.Fatal("seeded group count validation failed", "expected", totalGroups, "actual", len(byGroup))
	}

	groupIDs := make([]uuid.UUID, 0, len(byGroup))
	for id := range byGroup {
		groupIDs = append(groupIDs, id)
	}
	sort.Slice(groupIDs, func(i, j int) bool { return groupIDs[i].String() < groupIDs[j].String() })

	for _, groupID := range groupIDs {
		events := byGroup[groupID]
		for i := 1; i < len(events); i++ {
			expected := events[i-1].StartTime.Add(time.Duration(events[i-1].Duration) * time.Minute)
			if !events[i].StartTime.Equal(expected) {
				common.Fatal("adjacent timing validation failed", "group_id", groupID, "expected_start", expected, "actual_start", events[i].StartTime)
			}
		}
	}

	slog.Info("event seeding complete", "created_groups", createdGroups, "created_events", createdEvents, "past_start_groups", pastStartGroups)
}
