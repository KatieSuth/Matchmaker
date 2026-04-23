package model

// DiscordUser is the subset of the Discord /users/@me object we persist and display.
type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}
