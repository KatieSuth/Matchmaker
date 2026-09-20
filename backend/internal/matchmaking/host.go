package matchmaking

import "github.com/google/uuid"

// PickLobbyHost selects the lobby host from team players only: first willing volunteer
// by registration order, otherwise the earliest team player by registration order.
// Substitutes are never chosen — the host must be on a team so they can play in the custom.
func PickLobbyHost(lobby LobbyPlan) *uuid.UUID {
	var earliestVolunteer *Player
	var earliestMember *Player

	for i := range lobby.Roster {
		p := lobby.Roster[i]
		if earliestMember == nil || p.CreatedAt.Before(earliestMember.CreatedAt) {
			cp := p
			earliestMember = &cp
		}
		if p.CanLobbyHost {
			if earliestVolunteer == nil || p.CreatedAt.Before(earliestVolunteer.CreatedAt) {
				cp := p
				earliestVolunteer = &cp
			}
		}
	}

	if earliestVolunteer != nil {
		id := earliestVolunteer.UserID
		return &id
	}
	if earliestMember != nil {
		id := earliestMember.UserID
		return &id
	}
	return nil
}
