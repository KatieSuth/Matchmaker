package matchmaking

import (
	"time"

	"github.com/google/uuid"
)

// Settings holds env-backed fairness thresholds (injected from main via handler).
type Settings struct {
	FairnessOutlierGap         int
	FairnessTeamSeparation     float64
	FairnessReferenceTierCount int
}

// Config is per-game matchmaking input derived from event group + game mode.
type Config struct {
	EventID    uuid.UUID
	TeamSize   int
	SubMin     int
	SortLogic  string
	TierCount  int
	GameLabel  string
	LobbyCount int
	Slots      int
}

// Player is one registrant eligible for matchmaking (complete rank profile required).
type Player struct {
	UserID              uuid.UUID
	DiscordName         string
	DuoRequest          *string
	AvgRank             float64
	CanSubstitute       bool
	CanLobbyHost        bool
	RegisteredGameCount int
	CreatedAt           time.Time
	TeamNumber          *int
}

// LobbyPlan is the planned assignment for one lobby before persistence.
type LobbyPlan struct {
	Roster          []Player
	Subs            []Player
	HostID          *uuid.UUID
	FairnessWarning bool
}

// GamePlan is the full plan for one event (game) row.
type GamePlan struct {
	EventID             uuid.UUID
	Lobbies             []LobbyPlan
	SubCapacityAdjusted bool
}

// StrategyFunc assigns roster players to lobbies before shared post-processing.
type StrategyFunc func(players []Player, lobbyCount, slotsPerLobby int) []LobbyPlan
