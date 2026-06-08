// Package model_test table-tests mapping helpers from internal/db row types to API models.
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
// MapDbGamesToGames / MapDbGameToGame
// ============================================================

func TestMapDbGamesToGames_Empty(t *testing.T) {
	result := model.MapDbGamesToGames([]db.Game{})
	assert.Empty(t, result)
}

func TestMapDbGamesToGames_MapsAllFields(t *testing.T) {
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

	result := model.MapDbGamesToGames(input)

	require.Len(t, result, 1)
	assert.Equal(t, model.MapDbGameToGame(input[0]), result[0])
	assert.Equal(t, input[0].ID, result[0].ID)
	assert.Equal(t, input[0].Name, result[0].Name)
	assert.Equal(t, input[0].OwnerID, result[0].OwnerID)
	assert.Equal(t, input[0].CreatedAt, result[0].CreatedAt)
	assert.Equal(t, input[0].UpdatedAt, result[0].UpdatedAt)
}

func TestMapDbGamesToGames_NilOwnerID(t *testing.T) {
	input := []db.Game{
		{ID: uuid.New(), Name: "System Game", OwnerID: nil},
	}

	result := model.MapDbGamesToGames(input)

	require.Len(t, result, 1)
	assert.Nil(t, result[0].OwnerID)
}

func TestMapDbGamesToGames_MultipleGames(t *testing.T) {
	input := []db.Game{
		{ID: uuid.New(), Name: "Game A"},
		{ID: uuid.New(), Name: "Game B"},
		{ID: uuid.New(), Name: "Game C"},
	}

	result := model.MapDbGamesToGames(input)

	require.Len(t, result, len(input))
	for i, g := range result {
		assert.Equal(t, input[i].ID, g.ID)
		assert.Equal(t, input[i].Name, g.Name)
	}
}

// ============================================================
// MapDbGameModesToGameModes / MapDbGameModeToGameMode
// ============================================================

func TestMapDbGameModesToGameModes_Empty(t *testing.T) {
	result := model.MapDbGameModesToGameModes([]db.GameMode{})
	assert.Empty(t, result)
}

func TestMapDbGameModesToGameModes_MapsAllFields(t *testing.T) {
	gameID := uuid.New()
	ownerID := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	input := []db.GameMode{
		{
			ID:        uuid.New(),
			GameID:    gameID,
			Name:      "5v5",
			TeamSize:  5,
			OwnerID:   &ownerID,
			Duration:  45,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	result := model.MapDbGameModesToGameModes(input)

	require.Len(t, result, 1)
	assert.Equal(t, model.MapDbGameModeToGameMode(input[0]), result[0])
	assert.Equal(t, input[0].ID, result[0].ID)
	assert.Equal(t, input[0].GameID, result[0].GameID)
	assert.Equal(t, input[0].Name, result[0].Name)
	assert.Equal(t, input[0].TeamSize, result[0].TeamSize)
	assert.Equal(t, input[0].OwnerID, result[0].OwnerID)
	assert.Equal(t, input[0].Duration, result[0].Duration)
	assert.Equal(t, input[0].CreatedAt, result[0].CreatedAt)
	assert.Equal(t, input[0].UpdatedAt, result[0].UpdatedAt)
}

func TestMapDbGameModesToGameModes_NilOwnerID(t *testing.T) {
	gameID := uuid.New()
	input := []db.GameMode{
		{
			ID:       uuid.New(),
			GameID:   gameID,
			Name:     "Standard",
			TeamSize: 1,
			OwnerID:  nil,
			Duration: 30,
		},
	}

	result := model.MapDbGameModesToGameModes(input)

	require.Len(t, result, 1)
	assert.Nil(t, result[0].OwnerID)
}

func TestMapDbGameModesToGameModes_PreservesOrder(t *testing.T) {
	gameID := uuid.New()
	input := []db.GameMode{
		{ID: uuid.New(), GameID: gameID, Name: "Solo", TeamSize: 1, Duration: 20},
		{ID: uuid.New(), GameID: gameID, Name: "Duo", TeamSize: 2, Duration: 25},
		{ID: uuid.New(), GameID: gameID, Name: "Squad", TeamSize: 4, Duration: 30},
	}

	result := model.MapDbGameModesToGameModes(input)

	require.Len(t, result, 3)
	for i, m := range result {
		assert.Equal(t, input[i].ID, m.ID)
		assert.Equal(t, input[i].Name, m.Name)
		assert.Equal(t, input[i].TeamSize, m.TeamSize)
		assert.Equal(t, input[i].Duration, m.Duration)
	}
}

func TestMapDbGameModeToGameMode_MapsAllFields(t *testing.T) {
	gameID := uuid.New()
	ownerID := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	input := db.GameMode{
		ID:        uuid.New(),
		GameID:    gameID,
		Name:      "Competitive",
		TeamSize:  5,
		OwnerID:   &ownerID,
		Duration:  60,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := model.MapDbGameModeToGameMode(input)

	assert.Equal(t, input.ID, result.ID)
	assert.Equal(t, input.GameID, result.GameID)
	assert.Equal(t, input.Name, result.Name)
	assert.Equal(t, input.TeamSize, result.TeamSize)
	assert.Equal(t, input.OwnerID, result.OwnerID)
	assert.Equal(t, input.Duration, result.Duration)
	assert.Equal(t, input.CreatedAt, result.CreatedAt)
	assert.Equal(t, input.UpdatedAt, result.UpdatedAt)
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

// ============================================================
// Event group / registration mappings (event.go)
// ============================================================

func TestMapDbGetEventsForUserRowToDashboardEvent_MapsAllFields(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	eid := uuid.New()
	hid := uuid.New()
	input := db.GetEventsForUserRow{
		ID:               eid,
		GameName:         "Game",
		GameMode:         "5v5",
		EventDate:        now,
		HostID:           hid,
		HostName:         "host",
		RegisteredCount:  3,
		RegistrationOpen: true,
	}
	result := model.MapDbGetEventsForUserRowToDashboardEvent(input)
	assert.Equal(t, eid, result.ID)
	assert.Equal(t, "Game", result.GameName)
	assert.Equal(t, "5v5", result.GameMode)
	assert.Equal(t, now, result.EventDate)
	assert.Equal(t, hid, result.HostID)
	assert.Equal(t, "host", result.HostName)
	assert.Equal(t, 3, result.RegisteredCount)
	assert.True(t, result.RegistrationOpen)
}

func TestMapDbGetRegistrationDataByEventIdRowToEventRegistration_DiscordNamePointer(t *testing.T) {
	uid := uuid.New()
	eid := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	dn := "disco#1"
	input := db.GetRegistrationDataByEventIdRow{
		EventID:         eid,
		UserID:          uid,
		CanSubstitute:   true,
		CanLobbyHost:    false,
		DuoRequest:      nil,
		CreatedAt:       now,
		UpdatedAt:       now,
		DiscordName:     &dn,
		Pronouns:        "they/them",
		CurrentRankName: "Gold",
	}
	result := model.MapDbGetRegistrationDataByEventIdRowToEventRegistration(input)
	assert.Equal(t, dn, result.DiscordName)
	assert.Equal(t, "they/them", result.Pronouns)
	assert.Equal(t, "Gold", result.CurrentRankName)
}

func TestMapDbGetRegistrationDataByEventIdRowToEventRegistration_NilDiscordName(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	input := db.GetRegistrationDataByEventIdRow{
		EventID:     uuid.New(),
		UserID:      uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		DiscordName: nil,
	}
	result := model.MapDbGetRegistrationDataByEventIdRowToEventRegistration(input)
	assert.Equal(t, "", result.DiscordName)
}

func TestMapDbGetRegistrationDataByEventIdRowsToEventRegistrations_Empty(t *testing.T) {
	assert.Empty(t, model.MapDbGetRegistrationDataByEventIdRowsToEventRegistrations(nil))
	assert.Empty(t, model.MapDbGetRegistrationDataByEventIdRowsToEventRegistrations([]db.GetRegistrationDataByEventIdRow{}))
}

func TestMapDbGetRegistrationDataByEventIdRowsToEventRegistrations_MultipleRows(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	eid := uuid.New()
	u1 := uuid.New()
	u2 := uuid.New()
	dn := "player"
	rows := []db.GetRegistrationDataByEventIdRow{
		{
			EventID: eid, UserID: u1, CanSubstitute: true, CanLobbyHost: true,
			DuoRequest: nil, CreatedAt: now, UpdatedAt: now,
			DiscordName: &dn, Pronouns: "she/her", CurrentRankName: "Silver",
		},
		{
			EventID: eid, UserID: u2, CanSubstitute: false, CanLobbyHost: false,
			DuoRequest: nil, CreatedAt: now, UpdatedAt: now,
			DiscordName: nil, Pronouns: "", CurrentRankName: "Iron",
		},
	}
	got := model.MapDbGetRegistrationDataByEventIdRowsToEventRegistrations(rows)
	require.Len(t, got, 2)
	assert.Equal(t, model.MapDbGetRegistrationDataByEventIdRowToEventRegistration(rows[0]), got[0])
	assert.Equal(t, model.MapDbGetRegistrationDataByEventIdRowToEventRegistration(rows[1]), got[1])
}

func TestMapDbGetGroupEventsSummaryRowToEventGroupEvent_MapsNestedRegs(t *testing.T) {
	eid := uuid.New()
	gmid := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	summary := db.GetGroupEventsSummaryRow{
		ID:               eid,
		StartTime:        now,
		GameModeID:       gmid,
		GameModeName:     "Competitive",
		TeamSize:         5,
		RegisteredCount:  2,
		LobbiesCount:     0,
		PlayerRegistered: true,
	}
	regs := []model.EventRegistration{{UserID: uuid.New()}}
	result := model.MapDbGetGroupEventsSummaryRowToEventGroupEvent(summary, regs)
	assert.Equal(t, eid, result.ID)
	assert.Equal(t, now, result.StartTime)
	assert.Equal(t, gmid, result.GameModeID)
	assert.Equal(t, "Competitive", result.GameModeName)
	assert.Equal(t, 5, result.TeamSize)
	assert.Equal(t, 2, result.RegisteredCount)
	assert.Equal(t, 0, result.LobbiesCount)
	assert.True(t, result.PlayerRegistered)
	assert.Equal(t, regs, result.Registrations)
}

func TestMapDbGetEventGroupDetailByIdRowToEventGroupDetail_MapsAllFields(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	oid := uuid.New()
	gmid := uuid.New()
	gid := uuid.New()
	row := db.GetEventGroupDetailByIdRow{
		ID:               uuid.New(),
		OwnerID:          oid,
		OwnerName:        "owner",
		OwnerPronouns:    "they/them",
		GameModeID:       gmid,
		GameModeName:     "Mode",
		GameID:           gid,
		GameName:         "Valorant",
		TeamSize:         5,
		SubMin:           1,
		RegistrationOpen: false,
		Region:           "AMER",
		SortLogic:        "ranked",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	events := []model.EventGroupEvent{{ID: uuid.New()}}
	result := model.MapDbGetEventGroupDetailByIdRowToEventGroupDetail(row, events)
	assert.Equal(t, row.ID, result.ID)
	assert.Equal(t, oid, result.OwnerID)
	assert.Equal(t, "owner", result.OwnerName)
	assert.Equal(t, "they/them", result.OwnerPronouns)
	assert.Equal(t, gmid, result.GameModeID)
	assert.Equal(t, "Mode", result.GameModeName)
	assert.Equal(t, gid, result.GameID)
	assert.Equal(t, "Valorant", result.GameName)
	assert.Equal(t, 5, result.TeamSize)
	assert.Equal(t, 1, result.SubMin)
	assert.False(t, result.RegistrationOpen)
	assert.Equal(t, "AMER", result.Region)
	assert.Equal(t, "ranked", result.SortLogic)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, now, result.UpdatedAt)
	assert.Equal(t, events, result.Events)
}

func TestMapDbGetLobbiesForEventRowToEventLobby(t *testing.T) {
	hostID := uuid.New()
	eventID := uuid.New()
	row := db.GetLobbiesForEventRow{
		ID:                    uuid.New(),
		EventID:               &eventID,
		Host:                  &hostID,
		FairnessWarning:       true,
		FairnessWarningAtLock: true,
	}
	result := model.MapDbGetLobbiesForEventRowToEventLobby(row)
	assert.Equal(t, row.ID, result.ID)
	assert.Equal(t, &hostID, result.HostID)
	assert.True(t, result.FairnessWarning)
	assert.True(t, result.FairnessWarningAtLock)
	assert.Empty(t, result.Teams)
	assert.Empty(t, result.Subs)
}

func TestMapDbGetPlayersForLobbyRowToLobbyPlayer(t *testing.T) {
	name := "player-one"
	createdAt := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 2, 8, 30, 0, 0, time.UTC)
	duo := "TeammateName"
	row := db.GetPlayersForLobbyRow{
		UserID:          uuid.New(),
		TeamNumber:      ptrInt32(1),
		DiscordName:     &name,
		Pronouns:        "they/them",
		CurrentRankName:  "Gold 1",
		CurrentRankOrder: 12,
		PeakRankOrder:    15,
		CanSubstitute:    true,
		CanLobbyHost:     false,
		DuoRequest:       &duo,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	}
	result := model.MapDbGetPlayersForLobbyRowToLobbyPlayer(row)
	assert.Equal(t, row.UserID, result.UserID)
	assert.Equal(t, "player-one", result.DiscordName)
	assert.Equal(t, "they/them", result.Pronouns)
	assert.Equal(t, "Gold 1", result.CurrentRankName)
	assert.Equal(t, 12, result.CurrentRankOrder)
	assert.Equal(t, 15, result.PeakRankOrder)
	assert.True(t, result.CanSubstitute)
	assert.False(t, result.CanLobbyHost)
	require.NotNil(t, result.DuoRequest)
	assert.Equal(t, duo, *result.DuoRequest)
	assert.Equal(t, createdAt, result.CreatedAt)
	assert.Equal(t, updatedAt, result.UpdatedAt)
}

func ptrInt32(v int32) *int32 {
	return &v
}
