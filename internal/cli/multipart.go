package cli

import (
	"net/http"
	"net/url"
	"os"

	"github.com/voska/loby/internal/client"
	"github.com/voska/loby/internal/errfmt"
)

// openImageParts opens each non-empty path in fields, returning the multipart
// file parts plus a cleanup func that closes the underlying *os.File handles.
// Paths may use the @-prefix convention used elsewhere in the CLI.
func openImageParts(fields map[string]string) ([]client.FilePart, func(), error) {
	var (
		parts   []client.FilePart
		files   []*os.File
		cleanup = func() {
			for _, f := range files {
				_ = f.Close()
			}
		}
	)
	for name, path := range fields {
		if path == "" {
			continue
		}
		if len(path) > 1 && path[0] == '@' {
			path = path[1:]
		}
		f, err := os.Open(path) //nolint:gosec // path is a user-supplied CLI argument
		if err != nil {
			cleanup()
			return nil, func() {}, errfmt.Wrap(errfmt.UsageError, err)
		}
		files = append(files, f)
		parts = append(parts, client.FilePart{Field: name, Filename: path, Reader: f})
	}
	return parts, cleanup, nil
}

// execMultipart is the shared POST helper for endpoints that take
// multipart/form-data. Honors --dry-run, idempotency keys, and the standard
// output formatting.
func execMultipart(g *Globals, command, path string, form url.Values, files []client.FilePart, out any) error {
	if g.DryRun {
		preview := map[string]any{
			"method": http.MethodPost,
			"path":   path,
			"form":   form,
		}
		if len(files) > 0 {
			names := make([]string, 0, len(files))
			for _, p := range files {
				names = append(names, p.Field+":"+p.Filename)
			}
			preview["files"] = names
		}
		return g.Writer().Render(preview)
	}
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	idem := g.IdempotencyKey
	if idem == "" {
		k, err := client.IdempotencyKey(command, formToMap(form), nil, true)
		if err != nil {
			return err
		}
		idem = k
	}
	resp, err := cl.Do(g.Context(), &client.Request{
		Method:         http.MethodPost,
		Path:           path,
		Form:           form,
		Files:          files,
		Out:            out,
		IdempotencyKey: idem,
	})
	if err != nil {
		return err
	}
	if resp.Replayed {
		g.Writer().Notice("idempotent replay (Lob returned cached response)")
	}
	return g.Writer().Render(out)
}

func formToMap(f url.Values) map[string]string {
	out := make(map[string]string, len(f))
	for k, v := range f {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out
}
