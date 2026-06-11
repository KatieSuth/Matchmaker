package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
)

type cleanupSummary struct {
	GroupsDeleted  int
	GroupsNotFound int
	UsersDeleted   int
}

func runCleanup(seed *common.SeedContext) {
	games := loadSystemGames(seed)
	scenarios := allScenarios(games)

	var summary cleanupSummary
	for _, sc := range scenarios {
		groupID := scenarioGroupID(sc.Key, sc.SortLogic)
		tag, err := seed.Pool.Exec(seed.Ctx, `DELETE FROM event_groups WHERE id = $1`, groupID)
		if err != nil {
			common.Fatal("failed deleting event group", "group_id", groupID, "error", err)
		}
		if tag.RowsAffected() > 0 {
			summary.GroupsDeleted++
		} else {
			summary.GroupsNotFound++
		}
	}

	tag, err := seed.Pool.Exec(seed.Ctx, `
		DELETE FROM users
		WHERE discord_id LIKE 'mm_%'
	`)
	if err != nil {
		common.Fatal("failed deleting scenario users", "error", err)
	}
	summary.UsersDeleted = int(tag.RowsAffected())

	fmt.Fprintf(os.Stdout, "Matchmaking seed cleanup complete\n")
	fmt.Fprintf(os.Stdout, "  groups deleted:           %d\n", summary.GroupsDeleted)
	fmt.Fprintf(os.Stdout, "  groups not found:         %d\n", summary.GroupsNotFound)
	fmt.Fprintf(os.Stdout, "  scenario users deleted:   %d\n", summary.UsersDeleted)
	slog.Info("matchmaking cleanup complete",
		"groups_deleted", summary.GroupsDeleted,
		"groups_not_found", summary.GroupsNotFound,
		"users_deleted", summary.UsersDeleted)
}

func ensureNoPartialState(seed *common.SeedContext, g *systemGames) {
	partialSeedStateError(seed, g)
}
