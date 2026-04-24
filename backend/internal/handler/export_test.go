package handler

// SetLoginTestHooks overrides LoginHandler internals for deterministic tests.
func SetLoginTestHooks(
	h *Handler,
	generateState func() (string, error),
	encodeStateCookie func(name string, value interface{}) (string, error),
) {
	if generateState != nil {
		h.generateState = generateState
	}
	if encodeStateCookie != nil {
		h.encodeStateCookie = encodeStateCookie
	}
}
