package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// parseContentArg accepts an HTML/text body, a URL, a template ID, or @file.
// Text-extension files (.html, .htm, .txt, .csv, .md) are returned as strings
// since Lob's JSON endpoints accept inline HTML. Binary extensions (.pdf,
// .png, .jpg, .jpeg, .gif, .tiff, .webp) are base64-encoded as a data URI,
// which Lob's JSON endpoints accept for file fields. Other extensions fall
// back to base64 with a warning-ready data URI so corruption never goes
// silent.
func parseContentArg(s string) any {
	if s == "" {
		return nil
	}
	if !strings.HasPrefix(s, "@") {
		return s
	}
	path := s[1:]
	buf, err := os.ReadFile(path) //nolint:gosec // path is a user-supplied CLI argument
	if err != nil {
		return s
	}
	if isTextExt(path) {
		return string(buf)
	}
	mime := mimeForExt(path)
	encoded := base64.StdEncoding.EncodeToString(buf)
	return "data:" + mime + ";base64," + encoded
}

// isTextExt reports whether the path looks like an inline-text artifact (HTML,
// CSV, plain text) that Lob's JSON endpoints accept verbatim.
func isTextExt(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".htm", ".txt", ".md", ".csv", ".tsv", ".json":
		return true
	default:
		return false
	}
}

// mimeForExt returns a best-guess MIME type for binary file inputs. Used when
// emitting a data: URI to Lob's JSON endpoints.
func mimeForExt(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".tiff", ".tif":
		return "image/tiff"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// parseJSONArg accepts either inline JSON or @path-to-json and returns the
// decoded value (map, slice, or scalar).
func parseJSONArg(s string) (any, error) {
	if s == "" {
		return nil, nil
	}
	raw := []byte(s)
	if strings.HasPrefix(s, "@") {
		buf, err := os.ReadFile(s[1:]) //nolint:gosec // path is a user-supplied CLI argument
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
