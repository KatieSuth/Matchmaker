package matchmaking_test

import (
	"testing"

	"github.com/KatieSuth/MatchmakerAPI/internal/matchmaking"
	"github.com/stretchr/testify/assert"
)

func TestValidationError_Message(t *testing.T) {
	err := &matchmaking.ValidationError{Message: "Game 1 needs more players"}
	assert.Equal(t, "Game 1 needs more players", err.Error())
}
