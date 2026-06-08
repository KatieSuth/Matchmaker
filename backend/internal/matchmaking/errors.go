package matchmaking

import "errors"

var (
	ErrInsufficientPlayers     = errors.New("insufficient players")
	ErrInsufficientSubstitutes = errors.New("insufficient substitutes")
)

// ValidationError carries a client-facing message for capacity or constraint failures.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
