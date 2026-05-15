// Package output renders command results in the four modes loby supports:
// human (colored TTY), json, plain (TSV), and ndjson. Stdout is for parseable
// data; stderr is for human-only hints, progress, and errors.
package output

import (
	"io"
	"os"
)

// Mode is one of the four output formats. Default is Human; commands switch to
// NDJSON explicitly for streaming.
type Mode int

const (
	// Human is colored, table/key-value output for terminals.
	Human Mode = iota
	// JSON emits a single JSON document (indented).
	JSON
	// Plain emits tab-separated values, no color, no decoration.
	Plain
	// NDJSON emits one JSON document per line. Used for list/tail/stream.
	NDJSON
)

// Options resolves the active output mode from the global CLI flags and env.
// Precedence: explicit flag > env var > TTY auto-detection.
type Options struct {
	JSONFlag       bool
	PlainFlag      bool
	ResultsOnly    bool
	NoColor        bool
	Quiet          bool
	Select         string
	Stdout, Stderr io.Writer
}

// Resolve picks the active Mode given environment and TTY state. The Stdout
// writer is consulted only for TTY detection; data is always written through w.
func (o Options) Resolve() Mode {
	switch {
	case o.JSONFlag:
		return JSON
	case o.PlainFlag:
		return Plain
	case envTrue("LOBY_JSON"):
		return JSON
	case envTrue("LOBY_PLAIN"):
		return Plain
	case envTrue("LOBY_AUTO_JSON") && !isTerminal(o.Stdout):
		return JSON
	default:
		return Human
	}
}

// envTrue treats "1", "true", "yes", "on" as true. Anything else is false.
func envTrue(key string) bool {
	switch os.Getenv(key) {
	case "1", "true", "TRUE", "yes", "YES", "on", "ON":
		return true
	default:
		return false
	}
}
