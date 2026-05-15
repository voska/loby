package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/voska/loby/internal/client"
)

// execListSilent is like execList but does not render the response — the
// caller picks elements out and writes them as NDJSON itself.
func execListSilent(g *Globals, path string, q url.Values, out any) error {
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	if _, err := cl.Do(g.Context(), &client.Request{
		Method: http.MethodGet,
		Path:   path,
		Query:  q,
		Out:    out,
	}); err != nil {
		return err
	}
	return nil
}

// renderLine emits a single object as one JSON line on stdout. Used by
// streaming commands (tail).
func renderLine(g *Globals, v any) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if _, err := fmt.Fprintln(g.Stdout, string(buf)); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	return nil
}
