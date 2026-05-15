package cli

// AccountCmd implements GET /v1/accounts — the active account's profile,
// balance, and feature flags. Useful for agents verifying credentials.
type AccountCmd struct{}

// Run sends the request.
func (c *AccountCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execGet(g, "/accounts", &out)
}
