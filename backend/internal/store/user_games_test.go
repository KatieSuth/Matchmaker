package store_test

import (
	"context"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/KatieSuth/MatchmakerAPI/internal/test_util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedGame fetches the first system game, skipping the test if none are seeded.
func seedGame(t *testing.T, q *db.Queries) db.Game {
	t.Helper()
	games, err := q.GetSystemGames(context.Background())
	require.NoError(t, err)
	if len(games) == 0 {
		t.Skip("no system games seeded — skipping")
	}
	return games[0]
}

// seedGameRank fetches the first rank for a game, skipping the test if none exist.
func seedGameRank(t *testing.T, q *db.Queries, gameID uuid.UUID) db.GameRank {
	t.Helper()
	ranks, err := q.GetRanksForGame(context.Background(), &gameID)
	require.NoError(t, err)
	if len(ranks) == 0 {
		t.Skip("no ranks seeded for game — skipping")
	}
	return ranks[0]
}

// ============================================================
// UpsertGameForUser — validation (real DB needed to check game/rank existence)
// ============================================================

func TestUpsertGameForUser_InvalidGame(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug1", "uguser1")
		rankID := uuid.New()
		inGameName := "player1"

		_, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      uuid.New(), // nonexistent game
			CurrentRank: &rankID,
			PeakRank:    &rankID,
			InGameName:  &inGameName,
		})
		assert.ErrorIs(t, err, store.ErrInvalidGame)
	})
}

func TestUpsertGameForUser_NilCurrentRank(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug2", "uguser2")
		game := seedGame(t, q)
		peakRank := uuid.New()
		inGameName := "player2"

		_, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: nil,
			PeakRank:    &peakRank,
			InGameName:  &inGameName,
		})
		assert.ErrorIs(t, err, store.ErrCurrentRankMissing)
	})
}

func TestUpsertGameForUser_NilPeakRank(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug3", "uguser3")
		game := seedGame(t, q)
		currentRank := uuid.New()
		inGameName := "player3"

		_, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &currentRank,
			PeakRank:    nil,
			InGameName:  &inGameName,
		})
		assert.ErrorIs(t, err, store.ErrPeakRankMissing)
	})
}

func TestUpsertGameForUser_UuidNilCurrentRank(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug4", "uguser4")
		game := seedGame(t, q)
		nilRank := uuid.Nil
		peakRank := uuid.New()
		inGameName := "player4"

		_, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &nilRank,
			PeakRank:    &peakRank,
			InGameName:  &inGameName,
		})
		assert.ErrorIs(t, err, store.ErrCurrentRankMissing)
	})
}

func TestUpsertGameForUser_InvalidCurrentRank(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug5", "uguser5")
		game := seedGame(t, q)
		nonexistentRank := uuid.New()
		inGameName := "player5"

		_, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &nonexistentRank,
			PeakRank:    &nonexistentRank,
			InGameName:  &inGameName,
		})
		assert.ErrorIs(t, err, store.ErrInvalidCurrentRank)
	})
}

func TestUpsertGameForUser_CurrentRankHigherThanPeakRank(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug6", "uguser6")
		game := seedGame(t, q)
		ranks, err := q.GetRanksForGame(context.Background(), &game.ID)
		require.NoError(t, err)
		if len(ranks) < 2 {
			t.Skip("need at least 2 ranks to test rank ordering — skipping")
		}

		// Find a higher-order rank and lower-order rank.
		var lowRank, highRank db.GameRank
		for _, r := range ranks {
			if lowRank.ID == uuid.Nil || r.Order < lowRank.Order {
				lowRank = r
			}
			if highRank.ID == uuid.Nil || r.Order > highRank.Order {
				highRank = r
			}
		}

		inGameName := "player6"
		// Set current rank higher than peak rank — should be rejected.
		_, err = s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &highRank.ID,
			PeakRank:    &lowRank.ID,
			InGameName:  &inGameName,
		})
		assert.ErrorIs(t, err, store.ErrInvalidRankOrder)
	})
}

// ============================================================
// UpsertGameForUser — create and update paths
// ============================================================

func TestUpsertGameForUser_CreatesNewEntry(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug7", "uguser7")
		game := seedGame(t, q)
		rank := seedGameRank(t, q, game.ID)
		inGameName := "player7"

		result, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &rank.ID,
			PeakRank:    &rank.ID,
			InGameName:  &inGameName,
			ShowRank:    true,
		})
		require.NoError(t, err)
		assert.Equal(t, game.ID, result.GameID)
		assert.Equal(t, seeded.ID, result.UserID)
	})
}

func TestUpsertGameForUser_UpdatesExistingEntry(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug8", "uguser8")
		game := seedGame(t, q)
		rank := seedGameRank(t, q, game.ID)
		inGameName := "player8"
		updatedName := "player8-updated"

		// Create.
		_, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &rank.ID,
			PeakRank:    &rank.ID,
			InGameName:  &inGameName,
		})
		require.NoError(t, err)

		// Update.
		result, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &rank.ID,
			PeakRank:    &rank.ID,
			InGameName:  &updatedName,
			ShowRank:    true,
		})
		require.NoError(t, err)
		assert.Equal(t, &updatedName, result.InGameName)
	})
}

// ============================================================
// GetUserGamesForUser
// ============================================================

func TestGetUserGamesForUser_EmptyForNewUser(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug9", "uguser9")

		games, err := s.GetUserGamesForUser(context.Background(), seeded.ID)
		require.NoError(t, err)
		assert.Empty(t, games)
	})
}

func TestGetUserGamesForUser_ReturnsAddedGames(t *testing.T) {
	test_util.WithTestTx(t, func(q *db.Queries, s *store.PostgresStore) {
		seeded := seedUser(t, q, "discord-ug10", "uguser10")
		game := seedGame(t, q)
		rank := seedGameRank(t, q, game.ID)
		inGameName := "player10"

		_, err := s.UpsertGameForUser(context.Background(), seeded.ID, model.UserGame{
			GameID:      game.ID,
			CurrentRank: &rank.ID,
			PeakRank:    &rank.ID,
			InGameName:  &inGameName,
		})
		require.NoError(t, err)

		games, err := s.GetUserGamesForUser(context.Background(), seeded.ID)
		require.NoError(t, err)
		require.Len(t, games, 1)
		assert.Equal(t, game.ID, games[0].GameID)
	})
}
