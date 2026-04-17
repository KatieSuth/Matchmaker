package model_test

import (
	"testing"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// MapDbGameToGame
// ============================================================

func TestMapDbGameToGame_Empty(t *testing.T) {
	result := model.MapDbGameToGame([]db.Game{})
	assert.Empty(t, result)
}

func TestMapDbGameToGame_MapsAllFields(t *testing.T) {
	ownerID := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	input := []db.Game{
		{
			ID:        uuid.New(),
			Name:      "Valorant",
			OwnerID:   &ownerID,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	result := model.MapDbGameToGame(input)

	require.Len(t, result, 1)
	assert.Equal(t, input[0].ID, result[0].ID)
	assert.Equal(t, input[0].Name, result[0].Name)
	assert.Equal(t, input[0].OwnerID, result[0].OwnerID)
	assert.Equal(t, input[0].CreatedAt, result[0].CreatedAt)
	assert.Equal(t, input[0].UpdatedAt, result[0].UpdatedAt)
}

func TestMapDbGameToGame_NilOwnerID(t *testing.T) {
	input := []db.Game{
		{ID: uuid.New(), Name: "System Game", OwnerID: nil},
	}

	result := model.MapDbGameToGame(input)

	require.Len(t, result, 1)
	assert.Nil(t, result[0].OwnerID)
}

func TestMapDbGameToGame_MultipleGames(t *testing.T) {
	input := []db.Game{
		{ID: uuid.New(), Name: "Game A"},
		{ID: uuid.New(), Name: "Game B"},
		{ID: uuid.New(), Name: "Game C"},
	}

	result := model.MapDbGameToGame(input)

	require.Len(t, result, len(input))
	for i, g := range result {
		assert.Equal(t, input[i].ID, g.ID)
		assert.Equal(t, input[i].Name, g.Name)
	}
}

// ============================================================
// MapDbGameRanksToGameRanks
// ============================================================

func TestMapDbGameRanksToGameRanks_Empty(t *testing.T) {
	result := model.MapDbGameRanksToGameRanks([]db.GameRank{})
	assert.Empty(t, result)
}

func TestMapDbGameRanksToGameRanks_MapsAllFields(t *testing.T) {
	gameID := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	input := []db.GameRank{
		{
			ID:        uuid.New(),
			GameID:    &gameID,
			Name:      "Bronze",
			Order:     1,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	result := model.MapDbGameRanksToGameRanks(input)

	require.Len(t, result, 1)
	assert.Equal(t, input[0].ID, result[0].ID)
	assert.Equal(t, input[0].GameID, result[0].GameID)
	assert.Equal(t, input[0].Name, result[0].Name)
	assert.Equal(t, input[0].Order, result[0].Order)
	assert.Equal(t, input[0].CreatedAt, result[0].CreatedAt)
	assert.Equal(t, input[0].UpdatedAt, result[0].UpdatedAt)
}

func TestMapDbGameRanksToGameRanks_NilGameID(t *testing.T) {
	input := []db.GameRank{
		{ID: uuid.New(), GameID: nil, Name: "Unranked", Order: 0},
	}

	result := model.MapDbGameRanksToGameRanks(input)

	require.Len(t, result, 1)
	assert.Nil(t, result[0].GameID)
}

func TestMapDbGameRanksToGameRanks_PreservesOrder(t *testing.T) {
	gameID := uuid.New()
	input := []db.GameRank{
		{ID: uuid.New(), GameID: &gameID, Name: "Bronze", Order: 1},
		{ID: uuid.New(), GameID: &gameID, Name: "Silver", Order: 2},
		{ID: uuid.New(), GameID: &gameID, Name: "Gold", Order: 3},
	}

	result := model.MapDbGameRanksToGameRanks(input)

	require.Len(t, result, 3)
	for i, r := range result {
		assert.Equal(t, input[i].Name, r.Name)
		assert.Equal(t, input[i].Order, r.Order)
	}
}

// ============================================================
// MapDbGameRankToGameRank (singular)
// ============================================================

func TestMapDbGameRankToGameRank_MapsAllFields(t *testing.T) {
	gameID := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	input := db.GameRank{
		ID:        uuid.New(),
		GameID:    &gameID,
		Name:      "Diamond",
		Order:     4,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := model.MapDbGameRankToGameRank(input)

	assert.Equal(t, input.ID, result.ID)
	assert.Equal(t, input.GameID, result.GameID)
	assert.Equal(t, input.Name, result.Name)
	assert.Equal(t, input.Order, result.Order)
	assert.Equal(t, input.CreatedAt, result.CreatedAt)
	assert.Equal(t, input.UpdatedAt, result.UpdatedAt)
}

// ============================================================
// MapDbRefreshTokenToRefreshToken
// ============================================================

func TestMapDbRefreshTokenToRefreshToken_MapsAllFields(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	input := db.RefreshToken{
		Token:     "abc123",
		UserID:    uuid.New(),
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := model.MapDbRefreshTokenToRefreshToken(input)

	assert.Equal(t, input.Token, result.Token)
	assert.Equal(t, input.UserID, result.UserID)
	assert.Equal(t, input.ExpiresAt, result.ExpiresAt)
	assert.Equal(t, input.CreatedAt, result.CreatedAt)
	assert.Equal(t, input.UpdatedAt, result.UpdatedAt)
}

// ============================================================
// MapDbUserToUser
// ============================================================

func TestMapDbUserToUser_MapsAllFields(t *testing.T) {
	discordID := "discord-123"
	discordName := "testuser"
	imageURL := "https://cdn.discordapp.com/avatars/abc.png"
	pronouns := "they/them"
	region := "EU"
	now := time.Now().Truncate(time.Millisecond)

	input := db.User{
		ID:           uuid.New(),
		DiscordID:    &discordID,
		DiscordName:  &discordName,
		ImageUrl:     &imageURL,
		Pronouns:     &pronouns,
		ShowPronouns: true,
		Region:       &region,
		NewUser:      false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := model.MapDbUserToUser(input)

	assert.Equal(t, input.ID, result.ID)
	assert.Equal(t, input.DiscordID, result.DiscordID)
	assert.Equal(t, input.DiscordName, result.DiscordName)
	assert.Equal(t, input.ImageUrl, result.ImageUrl)
	assert.Equal(t, input.Pronouns, result.Pronouns)
	assert.Equal(t, input.ShowPronouns, result.ShowPronouns)
	assert.Equal(t, input.Region, result.Region)
	assert.Equal(t, input.NewUser, result.NewUser)
	assert.Equal(t, input.CreatedAt, result.CreatedAt)
	assert.Equal(t, input.UpdatedAt, result.UpdatedAt)
}

func TestMapDbUserToUser_NilOptionalFields(t *testing.T) {
	input := db.User{
		ID:          uuid.New(),
		DiscordID:   nil,
		DiscordName: nil,
		ImageUrl:    nil,
		Pronouns:    nil,
		Region:      nil,
	}

	result := model.MapDbUserToUser(input)

	assert.Nil(t, result.DiscordID)
	assert.Nil(t, result.DiscordName)
	assert.Nil(t, result.ImageUrl)
	assert.Nil(t, result.Pronouns)
	assert.Nil(t, result.Region)
}

// ============================================================
// MapDbUserGamesToUserGames
// ============================================================

func TestMapDbUserGamesToUserGames_Empty(t *testing.T) {
	result := model.MapDbUserGamesToUserGames([]db.GetGamesForUserRow{})
	assert.Empty(t, result)
}

func TestMapDbUserGamesToUserGames_MapsAllFields(t *testing.T) {
	currentRank := uuid.New()
	peakRank := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	inGameName := "ProPlayer99"

	input := []db.GetGamesForUserRow{
		{
			UserID:          uuid.New(),
			GameID:          uuid.New(),
			GameName:        "Valorant",
			InGameName:      inGameName,
			CurrentRank:     &currentRank,
			CurrentRankName: "Diamond",
			PeakRank:        &peakRank,
			PeakRankName:    "Radiant",
			ShowRank:        true,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	result := model.MapDbUserGamesToUserGames(input)

	require.Len(t, result, 1)
	assert.Equal(t, input[0].UserID, result[0].UserID)
	assert.Equal(t, input[0].GameID, result[0].GameID)
	assert.Equal(t, input[0].GameName, result[0].GameName)
	assert.Equal(t, &inGameName, result[0].InGameName)
	assert.Equal(t, input[0].CurrentRank, result[0].CurrentRank)
	assert.Equal(t, input[0].CurrentRankName, result[0].CurrentRankName)
	assert.Equal(t, input[0].PeakRank, result[0].PeakRank)
	assert.Equal(t, input[0].PeakRankName, result[0].PeakRankName)
	assert.Equal(t, input[0].ShowRank, result[0].ShowRank)
	assert.Equal(t, input[0].CreatedAt, result[0].CreatedAt)
	assert.Equal(t, input[0].UpdatedAt, result[0].UpdatedAt)
}

func TestMapDbUserGamesToUserGames_NilRanks(t *testing.T) {
	input := []db.GetGamesForUserRow{
		{
			UserID:      uuid.New(),
			GameID:      uuid.New(),
			GameName:    "Apex Legends",
			CurrentRank: nil,
			PeakRank:    nil,
		},
	}

	result := model.MapDbUserGamesToUserGames(input)

	require.Len(t, result, 1)
	assert.Nil(t, result[0].CurrentRank)
	assert.Nil(t, result[0].PeakRank)
}

// ============================================================
// MapDbUserGameToUserGame (singular)
// ============================================================

func TestMapDbUserGameToUserGame_MapsAllFields(t *testing.T) {
	currentRank := uuid.New()
	peakRank := uuid.New()
	inGameName := "TopFragger"
	now := time.Now().Truncate(time.Millisecond)

	input := db.UserGame{
		UserID:      uuid.New(),
		GameID:      uuid.New(),
		InGameName:  inGameName,
		CurrentRank: &currentRank,
		PeakRank:    &peakRank,
		ShowRank:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := model.MapDbUserGameToUserGame(input)

	assert.Equal(t, input.UserID, result.UserID)
	assert.Equal(t, input.GameID, result.GameID)
	assert.Equal(t, &inGameName, result.InGameName)
	assert.Equal(t, input.CurrentRank, result.CurrentRank)
	assert.Equal(t, input.PeakRank, result.PeakRank)
	assert.Equal(t, input.ShowRank, result.ShowRank)
	assert.Equal(t, input.CreatedAt, result.CreatedAt)
	assert.Equal(t, input.UpdatedAt, result.UpdatedAt)
}
