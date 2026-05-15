package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// parseAddressArg accepts one of:
//   - empty string → nil
//   - address ID (adr_…) → string (Lob inflates from ID)
//   - JSON object → map[string]any
//   - @path → contents of file as JSON object
//
// Lob's API accepts either an ID or an inline object for to/from fields.
func parseAddressArg(s string) any {
	if s == "" {
		return nil
	}
	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "@") {
		v, err := parseJSONArg(s)
		if err != nil {
			return s // fall back: let Lob reject if malformed
		}
		return v
	}
	return s
}

// parseContentArg accepts an HTML string, a URL, a template ID, or @file.html.
// Returns the raw string for IDs/URLs and the file contents for @path.
func parseContentArg(s string) any {
	if s == "" {
		return nil
	}
	if strings.HasPrefix(s, "@") {
		buf, err := os.ReadFile(s[1:])
		if err != nil {
			return s
		}
		return string(buf)
	}
	return s
}

// parseJSONArg accepts either inline JSON or @path-to-json and returns the
// decoded value (map, slice, or scalar).
func parseJSONArg(s string) (any, error) {
	if s == "" {
		return nil, nil
	}
	raw := []byte(s)
	if strings.HasPrefix(s, "@") {
		buf, err := os.ReadFile(s[1:])
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", s[1:], err)
		}
		raw = buf
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	return out, nil
}
