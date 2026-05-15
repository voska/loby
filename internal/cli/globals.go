// Package cli wires loby's command tree using Kong. Every command receives a
// *Globals via the Kong context and produces its output through an
// *output.Writer constructed from the globals.
package cli

import (
	"context"
	"io"
	"os"

	"github.com/voska/loby/internal/output"
)

// Globals holds the flags every command observes. They sit at the root level
// of the Kong struct and propagate to subcommands via inheritance.
type Globals struct {
	JSON           bool   `short:"j" help:"Emit JSON to stdout."                                        env:"LOBY_JSON"`
	Plain          bool   `short:"p" help:"Emit tab-separated values to stdout."                        env:"LOBY_PLAIN"`
	ResultsOnly    bool   `          help:"Strip the metadata envelope; emit data only."`
	Select         string `          help:"Project output to a subset of fields (comma-separated, dot-paths). e.g. --select id,to.city"`
	NoColor        bool   `          help:"Disable ANSI color in human output."                          env:"NO_COLOR"`
	Quiet          bool   `short:"q" help:"Suppress stderr hints; emit bare values."`
	NoInput        bool   `          help:"Fail rather than prompt for input. Implied when stdin is not a TTY."`
	APIKey         string `          help:"Lob API key (overrides keyring + env)."                       env:"LOB_API_KEY"`
	Profile        string `          help:"Named auth profile."                                          env:"LOB_PROFILE"  default:"default"`
	DryRun         bool   `short:"n" help:"Preview the request body as JSON; do not execute mutations."`
	IdempotencyKey string `          help:"Override the auto-generated Idempotency-Key for create operations."`
	Debug          bool   `          help:"Verbose logging to stderr."                                  env:"LOBY_DEBUG"`

	// Injected at runtime, not parsed from the command line.
	Stdout io.Writer       `kong:"-"`
	Stderr io.Writer       `kong:"-"`
	Stdin  io.Reader       `kong:"-"`
	Ctx    context.Context `kong:"-"`
}

// Writer builds the output.Writer for the current command.
func (g *Globals) Writer() *output.Writer {
	stdout := g.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := g.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	return output.New(output.Options{
		JSONFlag:    g.JSON,
		PlainFlag:   g.Plain,
		ResultsOnly: g.ResultsOnly,
		NoColor:     g.NoColor,
		Quiet:       g.Quiet,
		Select:      g.Select,
		Stdout:      stdout,
		Stderr:      stderr,
	})
}

// Context returns the command context, defaulting to context.Background().
func (g *Globals) Context() context.Context {
	if g.Ctx == nil {
		return context.Background()
	}
	return g.Ctx
}
