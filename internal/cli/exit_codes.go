package cli

import "github.com/voska/loby/internal/errfmt"

// ExitCodesCmd implements `loby exit-codes`.
type ExitCodesCmd struct{}

// Run dumps the canonical exit-code taxonomy.
func (c *ExitCodesCmd) Run(g *Globals) error {
	return g.Writer().Render(errfmt.Table)
}
