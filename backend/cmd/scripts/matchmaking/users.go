package main

import (
	"fmt"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
)

func createScenarioPlayer(
	seed *common.SeedContext,
	scenarioKey, sortLogic string,
	slot int,
	game *gameInfo,
	rankOrder int,
) (uuid.UUID, string) {
	userID := scenarioUserID(scenarioKey, sortLogic, slot)
	discordID := scenarioDiscordID(scenarioKey, sortLogic, slot)
	discordName := scenarioDiscordName(scenarioKey, sortLogic, slot)
	region := regionForSlot(slot)
	imageURL := fmt.Sprintf("mm_seed_image_%s", discordID)

	var exists bool
	err := seed.Pool.QueryRow(seed.Ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists)
	if err != nil {
		common.Fatal("failed checking scenario user", "user_id", userID, "error", err)
	}

	if !exists {
		_, err = seed.Pool.Exec(seed.Ctx, `
			INSERT INTO users (id, discord_id, discord_name, image_url, region, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`, userID, discordID, discordName, imageURL, region)
		if err != nil {
			common.Fatal("failed creating scenario user", "discord_id", discordID, "error", err)
		}
	}

	if userGameExists(seed, userID, game.ID) {
		return userID, discordName
	}

	rankID := game.rankIDForOrder(rankOrder)
	inGameName := fmt.Sprintf("MM_IGN_%s_%02d", scenarioKey, slot)
	_, err = seed.Queries.CreateGameForUser(seed.Ctx, db.CreateGameForUserParams{
		UserID:      userID,
		GameID:      game.ID,
		InGameName:  inGameName,
		CurrentRank: &rankID,
		PeakRank:    &rankID,
		AvgRank:     &rankID,
		ShowRank:    true,
	})
	if err != nil {
		common.Fatal("failed creating scenario user_game", "user_id", userID, "game", game.Name, "error", err)
	}

	return userID, discordName
}
