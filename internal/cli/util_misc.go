package cli

import "strconv"

// itoa is the strconv alias used throughout the CLI for small ints in query
// strings. Kept separate from arg_parsers to make grep-by-purpose easier.
func itoa(n int) string { return strconv.Itoa(n) }
