package main

import (
	"fmt"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/google/uuid"
)

// noShowHistoryGroupID is the index-th past one-game group holding a scenario's seeded no-show
// marks. Index 0 keeps the ID of the former single history group so cleanup still removes data
// seeded before no-shows were counted per event group.
func noShowHistoryGroupID(sc Scenario, index int) uuid.UUID {
	if index == 0 {
		return scenarioGroupID(sc.Key+"_history", sc.SortLogic)
	}
	return scenarioGroupID(fmt.Sprintf("%s_history_%d", sc.Key, index), sc.SortLogic)
}

// noShowHistoryGroupUseCase is the name suffix for the index-th history group.
func noShowHistoryGroupUseCase(sc Scenario, index int) string {
	return fmt.Sprintf("past no-shows %s #%d", sc.Key, index+1)
}

// scenarioHostSlot is the user slot of the seeded host in viewer scenarios; player slots stay below it.
const scenarioHostSlot = 99

// maxScenarioNoShows is the most no-shows any registrant in the scenario needs (including the
// --host user in a viewer scenario), which is also how many past history groups must exist.
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

// seedNoShowHistory creates closed past groups owned by ownerID, each holding one 5v5 game (one
// week apart, all inside the 6-month lookback). Each player is registered for, and marked absent
// in, the first RegSpec.NoShows of them; in a viewer scenario viewerID gets sc.ViewerNoShows the
// same way. One game per group because no-shows count once per event group. Players must
// already exist. No-op when nobody has no-shows.
func seedNoShowHistory(seed *common.SeedContext, ownerID, viewerID uuid.UUID, game *gameInfo, sc Scenario) {
	pastGroups := maxScenarioNoShows(sc)
	if pastGroups == 0 {
		return
	}

	mode := game.modeByName(modeName5v5)
	now := time.Now().UTC()
	eventIDs := make([]uuid.UUID, pastGroups)
	for i := range eventIDs {
		groupID := noShowHistoryGroupID(sc, i)
		name := scenarioGroupName(sc.SortLogic, noShowHistoryGroupUseCase(sc, i))
		insertEventGroup(seed, groupID, ownerID, 0, sc.Region, sc.SortLogic, name, 0, false)
		eventIDs[i] = scenarioEventID(groupID, 0)
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
