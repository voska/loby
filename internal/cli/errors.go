package cli

import (
	"errors"

	"github.com/voska/loby/internal/errfmt"
)

// errfmtUsage is the canonical wrap for "user passed bad flags" — every cmd
// uses this so exit codes stay consistent.
func errfmtUsage(msg string) error {
	return errfmt.Wrap(errfmt.UsageError, errors.New(msg))
}
