package matchmaking

// AssignBalanced fills each lobby from rank windows so each team can take one player
// per band. Lobbies are dealt in parallel: each gets 2 from a window before any lobby
// gets 4 from that window. Invalid ladder input falls back to a skill snake so planning
// cannot panic.
func AssignBalanced(players []Player, cfg Config) []LobbyPlan {
	lobbyCount := cfg.LobbyCount
	slotsPerLobby := cfg.Slots
	if slotsPerLobby <= 0 {
		// Modes omit slots; two teams of teamSize is the roster.
		slotsPerLobby = cfg.TeamSize * 2
	}
	if lobbyCount <= 0 {
		return nil
	}

	// Missing ladder dimensions cannot build windows; snake still produces a plan.
	if cfg.TeamSize <= 0 || cfg.TierCount <= 0 {
		return assignBalancedSnake(players, lobbyCount, slotsPerLobby)
	}

	windows := buildRankWindows(cfg.TierCount, cfg.TeamSize)
	if len(windows) == 0 {
		return assignBalancedSnake(players, lobbyCount, slotsPerLobby)
	}

	remaining := bucketPlayersByWindow(players, windows, cfg.TierCount)
	lobbies := make([]LobbyPlan, lobbyCount)
	for i := range lobbies {
		lobbies[i].Roster = make([]Player, 0, slotsPerLobby)
	}

	// Ideal pass: two from each window (one destined for each team) before extras.
	// Inner loop is per lobby, not "fill lobby 0 completely," so every lobby gets
	// a pair from a band before any lobby takes a second pair from it.
	for w := range windows {
		for i := 0; i < lobbyCount; i++ {
			if len(lobbies[i].Roster)+2 > slotsPerLobby {
				continue
			}
			if len(remaining[w]) < 2 {
				break
			}
			// Typical ranks of the band, not its highest leftover (that is how a
			// Radiant lands in a thin Gold window).
			picked, rest := pickClosestToMidpoint(remaining[w], 2, windows[w].midpoint)
			remaining[w] = rest
			lobbies[i].Roster = append(lobbies[i].Roster, picked...)
		}
	}

	// Short lobbies take extra pairs from the fullest remaining windows, one pair
	// per lobby per round so lobby 0 cannot drain a band before lobby 1 is dealt.
	for {
		progressed := false
		for i := 0; i < lobbyCount; i++ {
			need := slotsPerLobby - len(lobbies[i].Roster)
			if need <= 0 {
				continue
			}
			donor := fullestWindowIndex(remaining)
			if donor < 0 {
				break
			}
			// Keep dealing in pairs when possible so the later team split still
			// sees even counts per band.
			want := 2
			if need < 2 {
				want = need
			}
			if len(remaining[donor]) < want {
				want = len(remaining[donor])
			}
			if want <= 0 {
				continue
			}
			picked, rest := pickClosestToMidpoint(remaining[donor], want, windows[donor].midpoint)
			remaining[donor] = rest
			lobbies[i].Roster = append(lobbies[i].Roster, picked...)
			progressed = true
		}
		if !progressed {
			break
		}
	}

	return lobbies
}

// assignBalancedSnake is the panic-safe fallback when the rank ladder cannot be
// split into windows. It still fills every lobby to capacity instead of aborting.
func assignBalancedSnake(players []Player, lobbyCount, slotsPerLobby int) []LobbyPlan {
	// Always return one plan per requested lobby, even if counts are invalid,
	// so callers can range without a nil/empty special case.
	lobbies := make([]LobbyPlan, lobbyCount)
	for i := range lobbies {
		lobbies[i].Roster = make([]Player, 0, slotsPerLobby)
	}
	if lobbyCount <= 0 || slotsPerLobby <= 0 {
		return lobbies
	}

	// Only as many players as there are roster seats; the rest stay unplaced.
	needed := lobbyCount * slotsPerLobby
	pool := append([]Player(nil), players...)
	sortPlayersByRankAsc(pool)
	if len(pool) > needed {
		pool = pool[:needed]
	}

	// Lowest to highest, snaking across lobbies so adjacent ranks do not clump
	// in lobby 0 the way a round-robin-from-the-left deal would.
	for i, p := range pool {
		lobbyIdx := snakeLobbyIndex(i, lobbyCount)
		lobbies[lobbyIdx].Roster = append(lobbies[lobbyIdx].Roster, p)
	}
	return lobbies
}

// snakeLobbyIndex maps the i-th player in an ascending skill list onto a lobby
// by walking 0..n-1, then n-1..0. That keeps similar ranks from all landing in
// lobby 0. A non-positive lobbyCount returns 0 so callers can index safely.
func snakeLobbyIndex(i, lobbyCount int) int {
	if lobbyCount <= 0 {
		return 0
	}
	round := i / lobbyCount
	pos := i % lobbyCount
	if round%2 == 1 {
		pos = lobbyCount - 1 - pos
	}
	return pos
}
