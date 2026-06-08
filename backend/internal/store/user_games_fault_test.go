package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/KatieSuth/MatchmakerAPI/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validUserGamePayload(t *testing.T, s *store.PostgresStore, gameID uuid.UUID) model.UserGame {
	t.Helper()
	ctx := context.Background()
	ranks, err := s.GetGameRanks(ctx, &gameID)
	require.NoError(t, err)
	require.NotEmpty(t, ranks)
	rankID := ranks[0].ID
	inGameName := "fault-test-player"
	return model.UserGame{
		GameID:      gameID,
		CurrentRank: &rankID,
		PeakRank:    &rankID,
		InGameName:  &inGameName,
	}
}

func TestUpsertGameForUser_DatabaseErrors(t *testing.T) {
	injectedErr := errors.New("injected database failure")
	s, tx := createEventTestStoreTx(t)
	ctx := context.Background()
	user := createTestUser(t, ctx, s)
	games, err := s.GetSystemGames(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, games)
	payload := validUserGamePayload(t, s, games[0].ID)

	cases := []struct {
		name               string
		failOnQueryRowCall int
		wantContains       string
		setup              func()
	}{
		{
			name:               "game lookup fails",
			failOnQueryRowCall: 1,
			wantContains:       "looking up game by ID",
		},
		{
			name:               "current rank lookup fails",
			failOnQueryRowCall: 2,
			wantContains:       "looking up current rank by ID",
		},
		{
			name:               "peak rank lookup fails",
			failOnQueryRowCall: 3,
			wantContains:       "looking up peak rank by ID",
		},
		{
			name:               "user game lookup fails",
			failOnQueryRowCall: 4,
			wantContains:       "looking up user game",
		},
		{
			name:               "create user game fails",
			failOnQueryRowCall: 5,
			wantContains:       "creating game for user",
		},
		{
			name:               "update user game fails",
			failOnQueryRowCall: 5,
			wantContains:       "updated game for user",
			setup: func() {
				_, err := s.UpsertGameForUser(ctx, user.ID, payload)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup()
			}
			faulty := &faultInjectTx{
				DBTX:               tx,
				failOnQueryRowCall: tc.failOnQueryRowCall,
				injectedErr:        injectedErr,
			}
			fs := store.NewPostgresStoreFromDBTXForTest(faulty)

			_, err := fs.UpsertGameForUser(ctx, user.ID, payload)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantContains)
		})
	}
}
