package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"golang.org/x/term"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/version"
)

// isStdinTerminal returns true when stdin is attached to a terminal. Used to
// auto-imply --no-input when the CLI is invoked from a pipe or CI.
func isStdinTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Run is the single entry point invoked by main. It parses args, dispatches to
// the matched command, and returns the process exit code.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := &Root{Globals: Globals{Stdout: stdout, Stderr: stderr, Stdin: os.Stdin, Ctx: ctx}}

	parser, err := kong.New(
		root,
		kong.Name("loby"),
		kong.Description(description),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true, Summary: true}),
		kong.Vars{"version": version.Get().Version, "none": ""},
		kong.Writers(stdout, stderr),
	)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "fatal: build CLI parser:", err)
		return errfmt.GeneralError
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			_, _ = fmt.Fprintln(stderr, "error:", parseErr)
			_ = parseErr.Context.PrintUsage(false)
			return errfmt.UsageError
		}
		_, _ = fmt.Fprintln(stderr, "error:", err)
		return errfmt.UsageError
	}

	root.Stdout = stdout
	root.Stderr = stderr
	root.Ctx = ctx
	// --no-input is implied when stdin is not a TTY (piped from another
	// command, CI, etc.). Agents that explicitly pass --no-input still win.
	if !root.NoInput && !isStdinTerminal() {
		root.NoInput = true
	}

	if err := kctx.Run(&root.Globals, parser); err != nil {
		if !root.Quiet {
			_, _ = fmt.Fprintln(stderr, "error:", err)
		}
		return errfmt.ExitCodeOf(err)
	}
	return errfmt.Success
}

const description = `loby — canonical CLI for Lob (direct mail).

Send postcards, letters, checks, and self-mailers; verify addresses; manage
campaigns and templates. Built for humans and AI agents: structured JSON on
stdout, hints on stderr, stable exit codes, full schema introspection.

Quick start:
  loby auth login
  loby addresses verify "185 Berry St, San Francisco, CA"
  loby postcards create --to <addr_id> --front front.html --back back.html

Agent introspection:
  loby schema --json
  loby exit-codes --json
  loby version --json

Docs: https://lobycli.com`
