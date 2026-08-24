package matchmaking

import "fmt"

// RequiredPlayers returns the minimum registrant count for n lobbies.
func RequiredPlayers(n, slots, subMin int) int {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return slots
	}
	return n*slots + n*subMin
}

// MaxLobbies finds the largest lobby count this registrant pool can support
// while still meeting n × sub_min. Whether that n benches every can-sub is left
// to PlanEvent: keep it when the non-sub rosters are even, otherwise drop to n−1.
func MaxLobbies(registeredCount, substituteCount, slots, subMin int) int {
	max := 0
	for n := 1; ; n++ {
		if registeredCount < RequiredPlayers(n, slots, subMin) {
			break
		}
		if n >= 2 && substituteCount < n*subMin {
			break
		}
		max = n
	}
	return max
}

// MaxSubstituteEligibleOnRoster is how many sub-eligible players may be rostered when n >= 2.
func MaxSubstituteEligibleOnRoster(substituteCount, lobbyCount, subMin int) int {
	if lobbyCount < 2 {
		return substituteCount
	}
	reserved := lobbyCount * subMin
	if substituteCount <= reserved {
		return 0
	}
	return substituteCount - reserved
}

// ValidateCapacity checks whether at least one lobby can be formed.
func ValidateCapacity(registeredCount, substituteCount, slots, subMin int, gameLabel string) (int, error) {
	if registeredCount == 0 {
		return 0, nil
	}
	lobbyCount := MaxLobbies(registeredCount, substituteCount, slots, subMin)
	if lobbyCount < 1 {
		required := RequiredPlayers(1, slots, subMin)
		return 0, &ValidationError{
			Message: fmt.Sprintf("%s needs at least %d players but only %d are registered", gameLabel, required, registeredCount),
		}
	}
	if lobbyCount >= 2 && substituteCount < lobbyCount*subMin {
		return 0, &ValidationError{
			Message: fmt.Sprintf("%s needs at least %d substitute-eligible players for %d lobbies but only %d registered",
				gameLabel, lobbyCount*subMin, lobbyCount, substituteCount),
		}
	}
	return lobbyCount, nil
}
