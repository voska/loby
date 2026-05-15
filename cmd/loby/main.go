// Command loby is the canonical CLI for Lob — direct mail for humans and AI
// agents. The entry point stays thin: it wires signals and dispatches to
// internal/cli.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/voska/loby/internal/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	os.Exit(cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
