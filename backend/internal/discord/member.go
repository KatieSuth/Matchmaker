package discord

import "github.com/KatieSuth/MatchmakerAPI/internal/model"

// MemberOfAny reports whether the user belongs to at least one required Discord guild (OR).
func MemberOfAny(userGuilds []model.DiscordGuild, requiredIDs []string) bool {
	if len(userGuilds) == 0 || len(requiredIDs) == 0 {
		return false
	}
	want := make(map[string]struct{}, len(requiredIDs))
	for _, id := range requiredIDs {
		if id == "" {
			continue
		}
		want[id] = struct{}{}
	}
	if len(want) == 0 {
		return false
	}
	for _, g := range userGuilds {
		if _, ok := want[g.ID]; ok {
			return true
		}
	}
	return false
}
