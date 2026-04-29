// Command users seeds development users for local testing only.
// Run order: this script should run first before events and registrations.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"sort"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	seedUserCount = 30
)

var seedRegions = []string{"AMER", "EMEA", "APAC"}

var seedPronounOptions = []string{"he/him", "she/her", "they/them", "other"}

type seededGame struct {
	id    uuid.UUID
	name  string
	ranks []db.GameRank // sorted by Order ascending (lower order = lower tier)
}

func pronounsForSeedIndex(i int) *string {
	s := seedPronounOptions[(i-1)%len(seedPronounOptions)]
	return &s
}

// showPronounsForSeedIndex is true for ~4 in 5 users (every fifth user hides pronouns).
func showPronounsForSeedIndex(i int) bool {
	return i%5 != 0
}

// pickCurrentPeakRankIDs returns rank IDs where peak is at least as high as current (by game_ranks.order),
// with order difference at most 4 steps.
func pickCurrentPeakRankIDs(sorted []db.GameRank, userIdx, gameIdx int) (uuid.UUID, uuid.UUID) {
	n := len(sorted)
	if n == 0 {
		common.Fatal("pickCurrentPeakRankIDs: empty ranks")
	}
	if n == 1 {
		id := sorted[0].ID
		return id, id
	}
	seed := userIdx*31 + gameIdx*17 + 13
	currentIdx := (seed * 7) % n
	deltaMax := 4
	if rem := n - 1 - currentIdx; rem < deltaMax {
		deltaMax = rem
	}
	peakIdx := currentIdx + (seed % (deltaMax + 1))
	return sorted[currentIdx].ID, sorted[peakIdx].ID
}

func seedUpsertUserGames(seed *common.SeedContext, userID uuid.UUID, userIdx int, games []seededGame) {
	for gi, g := range games {
		if len(g.ranks) == 0 {
			common.Fatal("seeded game has no ranks", "game", g.name)
		}
		currentID, peakID := pickCurrentPeakRankIDs(g.ranks, userIdx, gi)
		inGameName := fmt.Sprintf("SeedIGN_%03d_%d", userIdx, gi)
		showRank := (userIdx+gi)%3 != 0

		_, err := seed.Queries.GetGameForUserByIds(seed.Ctx, db.GetGameForUserByIdsParams{
			UserID: userID,
			GameID: g.id,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = seed.Queries.CreateGameForUser(seed.Ctx, db.CreateGameForUserParams{
				UserID:      userID,
				GameID:      g.id,
				InGameName:  inGameName,
				CurrentRank: &currentID,
				PeakRank:    &peakID,
				ShowRank:    showRank,
			})
			if err != nil {
				common.Fatal("failed creating user_game", "user_idx", userIdx, "game", g.name, "error", err)
			}
			continue
		}
		if err != nil {
			common.Fatal("failed checking user_game", "user_idx", userIdx, "game", g.name, "error", err)
		}
		_, err = seed.Queries.UpdateGameForUser(seed.Ctx, db.UpdateGameForUserParams{
			InGameName:  inGameName,
			CurrentRank: &currentID,
			PeakRank:    &peakID,
			ShowRank:    showRank,
			UserID:      userID,
			GameID:      g.id,
		})
		if err != nil {
			common.Fatal("failed updating user_game", "user_idx", userIdx, "game", g.name, "error", err)
		}
	}
}

func loadSeedGames(seed *common.SeedContext) []seededGame {
	rows, err := seed.Pool.Query(seed.Ctx, `
		SELECT id, name
		FROM games
		WHERE owner_id IS NULL
		ORDER BY name ASC
	`)
	if err != nil {
		common.Fatal("failed querying games", "error", err)
	}
	defer rows.Close()

	var games []seededGame
	for rows.Next() {
		var g seededGame
		if err := rows.Scan(&g.id, &g.name); err != nil {
			common.Fatal("failed scanning game", "error", err)
		}
		gid := g.id
		ranks, err := seed.Queries.GetRanksForGame(seed.Ctx, &gid)
		if err != nil {
			common.Fatal("failed querying game ranks", "game", g.name, "error", err)
		}
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].Order < ranks[j].Order
		})
		g.ranks = ranks
		games = append(games, g)
	}
	if err := rows.Err(); err != nil {
		common.Fatal("failed iterating games", "error", err)
	}

	if len(games) == 0 {
		common.Fatal("no games found (owner_id IS NULL); run migrations/seeds for games first")
	}
	return games
}

func main() {
	seed := common.NewSeedContext()
	defer seed.Close()

	games := loadSeedGames(seed)

	created := 0
	existing := 0

	for i := 1; i <= seedUserCount; i++ {
		discordID := fmt.Sprintf("seed_user_%03d", i)
		discordName := fmt.Sprintf("SeedUser%03d", i)
		imageURL := fmt.Sprintf("seed_image_%03d", i)
		region := seedRegions[(i-1)%len(seedRegions)]
		pronouns := pronounsForSeedIndex(i)
		showPronouns := showPronounsForSeedIndex(i)

		user, err := seed.Queries.GetUserByDiscordID(seed.Ctx, &discordID)
		if err == nil {
			_, updateErr := seed.Queries.UpdateUser(seed.Ctx, db.UpdateUserParams{
				Pronouns:     pronouns,
				ShowPronouns: showPronouns,
				Region:       &region,
				ID:           user.ID,
			})
			if updateErr != nil {
				common.Fatal("failed updating existing seed user", "discord_id", discordID, "error", updateErr)
			}
			seedUpsertUserGames(seed, user.ID, i, games)
			existing++
			continue
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			common.Fatal("failed to check existing user", "discord_id", discordID, "error", err)
		}

		createdUser, err := seed.Queries.CreateUser(seed.Ctx, db.CreateUserParams{
			DiscordID:   &discordID,
			DiscordName: &discordName,
			ImageUrl:    &imageURL,
		})
		if err != nil {
			common.Fatal("failed creating user", "discord_id", discordID, "error", err)
		}
		_, err = seed.Queries.UpdateUser(seed.Ctx, db.UpdateUserParams{
			Pronouns:     pronouns,
			ShowPronouns: showPronouns,
			Region:       &region,
			ID:           createdUser.ID,
		})
		if err != nil {
			common.Fatal("failed setting profile for seeded user", "discord_id", discordID, "error", err)
		}

		seedUpsertUserGames(seed, createdUser.ID, i, games)

		created++
	}

	var total int
	if err := seed.Pool.QueryRow(seed.Ctx, `SELECT COUNT(*) FROM users WHERE discord_id LIKE 'seed_user_%'`).Scan(&total); err != nil {
		common.Fatal("failed verifying user seed count", "error", err)
	}
	if total < seedUserCount {
		common.Fatal("seeded users are incomplete", "expected_at_least", seedUserCount, "actual", total)
	}

	slog.Info("user seeding complete", "created", created, "existing", existing, "total_seed_users", total, "games_per_user", len(games))
}
