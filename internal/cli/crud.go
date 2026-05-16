package cli

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/voska/loby/internal/client"
	"github.com/voska/loby/internal/errfmt"
)

// listQuery builds the standard Lob pagination query.
func listQuery(limit int, before, after string, includeTotal bool, extra url.Values) url.Values {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if before != "" {
		q.Set("before", before)
	}
	if after != "" {
		q.Set("after", after)
	}
	if includeTotal {
		q.Set("include[]", "total_count")
	}
	for k, vs := range extra {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	return q
}

// requireConfirm returns a usage error unless one of confirm/force is set.
func requireConfirm(confirm, force bool) error {
	if confirm || force {
		return nil
	}
	return errfmt.Wrap(errfmt.UsageError, errors.New("--confirm (or --force) is required for destructive operations"))
}

// resourcePath joins base + id with a leading slash, validating id against
// agent-supplied junk (control chars, traversal, query injection).
func resourcePath(base, id string) (string, error) {
	if id == "" {
		return "", errfmt.Wrap(errfmt.UsageError, errors.New("resource ID is required"))
	}
	for _, r := range id {
		if r < 0x20 || r == '/' || r == '?' || r == '#' || r == '%' {
			return "", errfmt.Wrap(errfmt.UsageError, fmt.Errorf("invalid resource ID %q", id))
		}
	}
	return fmt.Sprintf("/%s/%s", base, id), nil
}

// execCancel sends DELETE /<resource>/<id> — Lob's actual cancel mechanism
// for mailer resources. (The spec exposes cancel as DELETE on the single
// resource, not as a /cancel sub-route.) Only letters, checks, and snap_packs
// expose this; postcards and self_mailers cannot be cancelled via the API.
func execCancel(g *Globals, resource, id string, confirm, force bool) error {
	if err := requireConfirm(confirm, force); err != nil {
		return err
	}
	path, err := resourcePath(resource, id)
	if err != nil {
		return err
	}
	if g.DryRun {
		return g.Writer().Render(map[string]any{"method": http.MethodDelete, "path": path})
	}
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	out := map[string]any{}
	if _, err := cl.Do(g.Context(), &client.Request{Method: http.MethodDelete, Path: path, Out: &out}); err != nil {
		return err
	}
	return g.Writer().Render(out)
}
