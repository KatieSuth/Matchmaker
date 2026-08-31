package discord_test

import (
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/discord"
	"github.com/KatieSuth/MatchmakerAPI/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMemberOfAny(t *testing.T) {
	guilds := []model.DiscordGuild{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Beta"},
	}
	tests := []struct {
		name     string
		user     []model.DiscordGuild
		required []string
		want     bool
	}{
		{name: "empty user list", required: []string{"1"}, want: false},
		{name: "empty required list", user: guilds, want: false},
		{name: "overlap", user: guilds, required: []string{"2", "9"}, want: true},
		{name: "no overlap", user: guilds, required: []string{"9"}, want: false},
		{name: "extra guilds on user", user: guilds, required: []string{"1"}, want: true},
		{name: "extra required ids", user: guilds, required: []string{"9", "1", "8"}, want: true},
		{name: "blank required ids ignored", user: guilds, required: []string{"", " "}, want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, discord.MemberOfAny(tc.user, tc.required))
		})
	}
}
