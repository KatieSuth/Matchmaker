package main

import (
	"time"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/google/uuid"
)

// noShowHistoryGroupID is the past-games group that holds a scenario's seeded no-show marks.
func noShowHistoryGroupID(sc Scenario) uuid.UUID {
	return scenarioGroupID(sc.Key+"_history", sc.SortLogic)
}

// scenarioHostSlot is the user slot of the seeded host in viewer scenarios; player slots stay below it.
const scenarioHostSlot = 99

// maxScenarioNoShows is the most no-shows any registrant in the scenario needs (including the
// --host user in a viewer scenario), which is also how many past games the history group must hold.
func maxScenarioNoShows(sc Scenario) int {
	max := sc.ViewerNoShows
	for _, ev := range sc.Events {
		for _, reg := range ev.Registrations {
			if reg.NoShows > max {
				max = reg.NoShows
			}
		}
	}
	return max
}

// seedNoShowHistory creates a closed group of past 5v5 games owned by ownerID (one week apart,
// all inside the 6-month lookback). Each player is registered for, and marked absent in, the
// first RegSpec.NoShows of them; in a viewer scenario viewerID gets sc.ViewerNoShows the same way.
// Players must already exist. No-op when nobody has no-shows.
func seedNoShowHistory(seed *common.SeedContext, ownerID, viewerID uuid.UUID, game *gameInfo, sc Scenario) {
	pastGames := maxScenarioNoShows(sc)
	if pastGames == 0 {
		return
	}

	groupID := noShowHistoryGroupID(sc)
	name := scenarioGroupName(sc.SortLogic, "past no-shows "+sc.Key)
	insertEventGroup(seed, groupID, ownerID, 0, sc.Region, sc.SortLogic, name, 0, false)

	mode := game.modeByName(modeName5v5)
	now := time.Now().UTC()
	eventIDs := make([]uuid.UUID, pastGames)
	for i := range eventIDs {
		eventIDs[i] = scenarioEventID(groupID, i)
		insertEvent(seed, eventIDs[i], groupID, mode.ID, now.AddDate(0, 0, -7*(i+1)))
	}

	absences := map[uuid.UUID]int{}
	for _, ev := range sc.Events {
		for _, reg := range ev.Registrations {
			if reg.NoShows > 0 {
				absences[scenarioUserID(sc.Key, sc.SortLogic, reg.Slot)] = reg.NoShows
			}
		}
	}
	if sc.ViewerNoShows > 0 {
		absences[viewerID] = sc.ViewerNoShows
	}
	for userID, n := range absences {
		for i := 0; i < n; i++ {
			markPastNoShow(seed, sc, eventIDs[i], userID, ownerID)
		}
	}
}

// markPastNoShow registers userID for a past history game and marks them absent there.
func markPastNoShow(seed *common.SeedContext, sc Scenario, eventID, userID, ownerID uuid.UUID) {
	registerUser(seed, sc, eventID, userID)
	_, err := seed.Pool.Exec(seed.Ctx, `
		INSERT INTO no_shows (event_id, user_id, host_id) VALUES ($1, $2, $3)
		ON CONFLICT (event_id, user_id) DO NOTHING
	`, eventID, userID, ownerID)
	if err != nil {
		common.Fatal("failed seeding no-show", "scenario", sc.Key, "user_id", userID, "error", err)
	}
}

// registerUser adds a plain registration (no sub, no lobby host, no duo) for an existing user.
func registerUser(seed *common.SeedContext, sc Scenario, eventID, userID uuid.UUID) {
	_, err := seed.Pool.Exec(seed.Ctx, `
		INSERT INTO registrations (event_id, user_id, can_substitute, can_lobby_host, created_at, updated_at)
		VALUES ($1, $2, false, false, NOW(), NOW())
		ON CONFLICT (event_id, user_id) DO NOTHING
	`, eventID, userID)
	if err != nil {
		common.Fatal("failed registering user", "scenario", sc.Key, "user_id", userID, "error", err)
	}
}

// createScenarioHost creates the seeded host for a viewer scenario, who owns its groups and
// is removed by cleanup with the other mm_ users.
func createScenarioHost(seed *common.SeedContext, sc Scenario, game *gameInfo) uuid.UUID {
	id, _ := createScenarioPlayer(seed, sc.Key, sc.SortLogic, scenarioHostSlot, game, 10)
	return id
}
