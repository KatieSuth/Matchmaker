package logger

import (
	"context"
	"log/slog"

	"github.com/KatieSuth/MatchmakerAPI/internal/middleware"
)

type requestIDHandler struct {
	slog.Handler
}

func (h requestIDHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := middleware.RequestIDFromContext(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func New(base slog.Handler) slog.Handler {
	return requestIDHandler{base}
}
