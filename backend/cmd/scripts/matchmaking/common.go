package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
	"github.com/KatieSuth/MatchmakerAPI/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	gameKeyValorant = "valorant"
	gameKeyLoL      = "lol"

	modeName5v5        = "5v5"
	modeName3v3Skirmish = "3v3 Skirmish"
)

var (
	matchmakingSeedNamespace = uuid.MustParse("018e0000-0000-7000-8000-000000000001")
	seedRegions              = []string{"AMER", "EMEA", "APAC"}
)

type modeInfo struct {
	ID       uuid.UUID
	Name     string
	TeamSize int32
	Duration int32
}

type gameInfo struct {
	ID    uuid.UUID
	Name  string
	Ranks []db.GameRank
	Modes map[string]modeInfo
}

type systemGames struct {
	Valorant gameInfo
	LoL      gameInfo
}

func (g *gameInfo) modeByName(name string) modeInfo {
	mode, ok := g.Modes[name]
	if !ok {
		common.Fatal("game mode not found", "game", g.Name, "mode", name)
	}
	return mode
}

func (g *gameInfo) rankIDForOrder(order int) uuid.UUID {
	if len(g.Ranks) == 0 {
		common.Fatal("game has no ranks", "game", g.Name)
	}
	if order <= int(g.Ranks[0].Order) {
		return g.Ranks[0].ID
	}
	max := g.Ranks[len(g.Ranks)-1]
	if order >= int(max.Order) {
		return max.ID
	}
	for _, r := range g.Ranks {
		if int(r.Order) == order {
			return r.ID
		}
	}
	common.Fatal("rank order not found", "game", g.Name, "order", order)
	return uuid.Nil
}

func loadModesForGame(seed *common.SeedContext, gameID uuid.UUID) map[string]modeInfo {
	rows, err := seed.Pool.Query(seed.Ctx, `
		SELECT id, name, team_size, duration
		FROM game_modes
		WHERE game_id = $1 AND owner_id IS NULL
		ORDER BY name ASC
	`, gameID)
	if err != nil {
		common.Fatal("failed querying game modes", "game_id", gameID, "error", err)
	}
	defer rows.Close()

	modes := make(map[string]modeInfo)
	for rows.Next() {
		var mode modeInfo
		if err := rows.Scan(&mode.ID, &mode.Name, &mode.TeamSize, &mode.Duration); err != nil {
			common.Fatal("failed scanning game mode", "error", err)
		}
		modes[mode.Name] = mode
	}
	if err := rows.Err(); err != nil {
		common.Fatal("failed iterating game modes", "error", err)
	}
	return modes
}

func loadSystemGames(seed *common.SeedContext) *systemGames {
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

	byName := map[string]gameInfo{}
	for rows.Next() {
		var g gameInfo
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			common.Fatal("failed scanning game", "error", err)
		}
		gid := g.ID
		ranks, err := seed.Queries.GetRanksForGame(seed.Ctx, &gid)
		if err != nil {
			common.Fatal("failed querying game ranks", "game", g.Name, "error", err)
		}
		sort.Slice(ranks, func(i, j int) bool { return ranks[i].Order < ranks[j].Order })
		g.Ranks = ranks
		g.Modes = loadModesForGame(seed, g.ID)
		if _, ok := g.Modes[modeName5v5]; !ok {
			common.Fatal("5v5 mode not found", "game", g.Name)
		}
		byName[g.Name] = g
	}
	if err := rows.Err(); err != nil {
		common.Fatal("failed iterating games", "error", err)
	}

	val, ok := byName["Valorant"]
	if !ok {
		common.Fatal("Valorant game not found; run migrations first")
	}
	lol, ok := byName["League of Legends"]
	if !ok {
		common.Fatal("League of Legends game not found; run migrations first")
	}
	return &systemGames{Valorant: val, LoL: lol}
}

func (sg *systemGames) gameForKey(key string) gameInfo {
	switch key {
	case gameKeyValorant:
		return sg.Valorant
	case gameKeyLoL:
		return sg.LoL
	default:
		common.Fatal("unknown game key", "key", key)
		return gameInfo{}
	}
}

func scenarioGroupID(key, sortLogic string) uuid.UUID {
	return uuid.NewSHA1(matchmakingSeedNamespace, []byte(fmt.Sprintf("group:%s:%s", key, sortLogic)))
}

func scenarioUserID(key, sortLogic string, slot int) uuid.UUID {
	return uuid.NewSHA1(matchmakingSeedNamespace, []byte(fmt.Sprintf("user:%s:%s:%02d", key, sortLogic, slot)))
}

func scenarioEventID(groupID uuid.UUID, eventIndex int) uuid.UUID {
	return uuid.NewSHA1(matchmakingSeedNamespace, []byte(fmt.Sprintf("event:%s:%d", groupID, eventIndex)))
}

func scenarioDiscordID(key, sortLogic string, slot int) string {
	return fmt.Sprintf("mm_%s_%s_%02d", key, sortLogic, slot)
}

func scenarioDiscordName(key, sortLogic string, slot int) string {
	return fmt.Sprintf("MM_%s_%s_%02d", key, sortLogic, slot)
}

func regionForSlot(slot int) string {
	return seedRegions[slot%len(seedRegions)]
}

func resolveHost(seed *common.SeedContext, discordName string) uuid.UUID {
	if discordName == "" {
		common.Fatal("--host is required")
	}
	user, err := seed.Queries.GetUserByName(seed.Ctx, &discordName)
	if err != nil {
		common.Fatal("host user not found", "discord_name", discordName, "error", err)
	}
	return user.ID
}

func groupExists(seed *common.SeedContext, groupID uuid.UUID) bool {
	var exists bool
	err := seed.Pool.QueryRow(seed.Ctx, `
		SELECT EXISTS(SELECT 1 FROM event_groups WHERE id = $1)
	`, groupID).Scan(&exists)
	if err != nil {
		common.Fatal("failed checking event group", "group_id", groupID, "error", err)
	}
	return exists
}

func insertEventGroup(seed *common.SeedContext, groupID, ownerID uuid.UUID, subMin int32, region, sortLogic string) {
	_, err := seed.Pool.Exec(seed.Ctx, `
		INSERT INTO event_groups (id, owner_id, sub_min, registration_open, region, sort_logic, created_at, updated_at)
		VALUES ($1, $2, $3, true, $4, $5, NOW(), NOW())
	`, groupID, ownerID, subMin, region, sortLogic)
	if err != nil {
		common.Fatal("failed creating event group", "group_id", groupID, "error", err)
	}
}

func insertEvent(seed *common.SeedContext, eventID, groupID, modeID uuid.UUID, start time.Time) {
	_, err := seed.Pool.Exec(seed.Ctx, `
		INSERT INTO events (id, group_id, start_time, game_mode_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, eventID, groupID, start.UTC(), modeID)
	if err != nil {
		common.Fatal("failed creating event", "event_id", eventID, "error", err)
	}
}

func firstEventStart(groupIndex int) time.Time {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		common.Fatal("failed loading timezone", "error", err)
	}
	now := time.Now().In(loc)
	y, m, d := now.Date()
	base := time.Date(y, m, d, 19, 0, 0, 0, loc)
	if base.Before(now) {
		base = base.AddDate(0, 0, 1)
	}
	return base.AddDate(0, 0, groupIndex).UTC()
}

func userGameExists(seed *common.SeedContext, userID, gameID uuid.UUID) bool {
	_, err := seed.Queries.GetGameForUserByIds(seed.Ctx, db.GetGameForUserByIdsParams{
		UserID: userID,
		GameID: gameID,
	})
	if err == nil {
		return true
	}
	if err == pgx.ErrNoRows {
		return false
	}
	common.Fatal("failed checking user_game", "user_id", userID, "game_id", gameID, "error", err)
	return false
}

func defaultModeName(modeName string) string {
	if modeName == "" {
		return modeName5v5
	}
	return modeName
}
