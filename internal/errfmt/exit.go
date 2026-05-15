// Package errfmt defines loby's exit-code taxonomy and the typed errors that
// surface through it. The taxonomy is the agent contract — never reuse codes.
package errfmt

import "errors"

// Canonical exit codes. Documented in AGENTS.md and exposed via
// `loby exit-codes --json`.
const (
	Success         = 0
	GeneralError    = 1
	UsageError      = 2
	Empty           = 3
	AuthRequired    = 4
	NotFound        = 5
	Forbidden       = 6
	RateLimited     = 7
	Retryable       = 8
	PaymentRequired = 9
	ConfigError     = 10
)

// Code is the public, stable description for one exit code, mirrored 1:1 by
// `loby exit-codes --json` output.
type Code struct {
	Code        int    `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Transient   bool   `json:"transient"`
}

// Table is the authoritative ordered list of every exit code loby can return.
// Order matters — agents render this as a reference table.
var Table = []Code{
	{Success, "success", "Operation succeeded.", false},
	{GeneralError, "error", "Unspecified failure. See stderr.", false},
	{UsageError, "usage", "Invalid arguments or flags. Run --help.", false},
	{Empty, "empty", "Operation succeeded but returned no results.", false},
	{AuthRequired, "auth_required", "Missing or invalid Lob API key. Run `loby auth login`.", false},
	{NotFound, "not_found", "Resource does not exist.", false},
	{Forbidden, "forbidden", "API key lacks permission for this resource.", false},
	{RateLimited, "rate_limited", "Lob API rate limit hit. Retry after Retry-After.", true},
	{Retryable, "retryable", "Transient error (timeout, 5xx, network). Safe to retry.", true},
	{PaymentRequired, "payment_required", "Lob account has insufficient funds or billing issue.", false},
	{ConfigError, "config_error", "Local configuration is invalid or unreadable.", false},
}

// Coded wraps an error with a specific exit code. main() unwraps with
// errors.As to set the process status.
type Coded struct {
	ExitCode int
	Err      error
}

func (e *Coded) Error() string { return e.Err.Error() }
func (e *Coded) Unwrap() error { return e.Err }

// Wrap pairs an error with a specific exit code. Returns nil if err is nil.
func Wrap(code int, err error) error {
	if err == nil {
		return nil
	}
	return &Coded{ExitCode: code, Err: err}
}

// ExitCodeOf walks the error chain to find the first *Coded and returns its
// exit code. Defaults to GeneralError. Returns Success for nil.
func ExitCodeOf(err error) int {
	if err == nil {
		return Success
	}
	var c *Coded
	if errors.As(err, &c) {
		return c.ExitCode
	}
	return GeneralError
}
