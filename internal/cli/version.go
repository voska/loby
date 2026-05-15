package cli

import (
	"fmt"

	"github.com/voska/loby/internal/output"
	"github.com/voska/loby/internal/version"
)

// VersionCmd implements `loby version`.
type VersionCmd struct{}

// Run prints the build info, either as a one-line summary (human) or as JSON.
func (c *VersionCmd) Run(g *Globals) error {
	info := version.Get()
	w := g.Writer()
	if w.Mode() == output.Human {
		_, _ = fmt.Fprintf(g.Stdout, "loby %s\ncommit %s\nbuilt  %s\n%s %s/%s\n",
			info.Version, info.Commit, info.Date, info.Go, info.OS, info.Arch)
		return nil
	}
	return w.Render(info)
}
