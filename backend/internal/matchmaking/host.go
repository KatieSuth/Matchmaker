package matchmaking

import "github.com/google/uuid"

// PickLobbyHost selects the lobby host: first willing volunteer by registration order,
// otherwise the first assigned lobby member by registration order.
func PickLobbyHost(lobby LobbyPlan) *uuid.UUID {
	var all []Player
	all = append(all, lobby.Roster...)
	all = append(all, lobby.Subs...)

	var earliestVolunteer *Player
	var earliestMember *Player

	for i := range all {
		p := all[i]
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
