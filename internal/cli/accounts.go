package cli

// AccountCmd implements GET /v1/accounts/credits_balance — the active
// account's remaining Lob Credits. Useful for agents verifying credentials
// and checking pre-paid balance before mailing.
type AccountCmd struct{}

// Run sends the request.
func (c *AccountCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execGet(g, "/accounts/credits_balance", &out)
}
