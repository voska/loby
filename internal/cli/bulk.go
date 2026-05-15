package cli

import (
	"net/url"
	"os"

	"github.com/voska/loby/internal/client"
	"github.com/voska/loby/internal/errfmt"
)

// BulkCmd groups the bulk verification operations. Each endpoint is async:
// submit returns a job ID, poll status until completed, then download results.
type BulkCmd struct {
	US   BulkUSCmd   `cmd:"" help:"Bulk US address verifications (synchronous JSON array)."`
	Intl BulkIntlCmd `cmd:"" help:"Bulk international address verifications (synchronous JSON array)."`
	CSV  BulkCSVCmd  `cmd:"" help:"CSV-based async US verification jobs."`
}

// BulkUSCmd posts up to 100 addresses to /v1/bulk/us_verifications.
type BulkUSCmd struct {
	Addresses string `help:"JSON array of address objects (or @file.json)." required:""`
	Case      string `help:"Case transformation." enum:"upper,proper,default" default:"default"`
}

// Run sends the request.
func (c *BulkUSCmd) Run(g *Globals) error {
	body, err := parseJSONArg(c.Addresses)
	if err != nil {
		return errfmt.Wrap(errfmt.UsageError, err)
	}
	q := url.Values{}
	if c.Case != "" && c.Case != "default" {
		q.Set("case", c.Case)
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "bulk_us_verifications", "/bulk/us_verifications", q, map[string]any{"addresses": body}, &out)
}

// BulkIntlCmd posts up to 100 addresses to /v1/bulk/intl_verifications.
type BulkIntlCmd struct {
	Addresses string `help:"JSON array of address objects (or @file.json)." required:""`
}

// Run sends the request.
func (c *BulkIntlCmd) Run(g *Globals) error {
	body, err := parseJSONArg(c.Addresses)
	if err != nil {
		return errfmt.Wrap(errfmt.UsageError, err)
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "bulk_intl_verifications", "/bulk/intl_verifications", url.Values{}, map[string]any{"addresses": body}, &out)
}

// BulkCSVCmd groups async CSV-based bulk verification jobs.
type BulkCSVCmd struct {
	Submit   BulkCSVSubmitCmd   `cmd:"" help:"Submit a CSV file for async US verification."`
	Status   BulkCSVStatusCmd   `cmd:"" help:"Check status of a CSV verification job."`
	Download BulkCSVDownloadCmd `cmd:"" help:"Download results of a completed CSV verification job."`
	List     BulkCSVListCmd     `cmd:"" help:"List CSV verification jobs."`
	Delete   BulkCSVDeleteCmd   `cmd:"" help:"Delete a CSV verification job."`
}

// BulkCSVSubmitCmd posts a CSV to /v1/us_verifications.
type BulkCSVSubmitCmd struct {
	Path        string `arg:"" help:"Path to CSV file with addresses."`
	Description string `help:"Internal description."`
}

// Run sends the multipart request.
func (c *BulkCSVSubmitCmd) Run(g *Globals) error {
	f, err := os.Open(c.Path) //nolint:gosec // user-supplied CLI argument
	if err != nil {
		return errfmt.Wrap(errfmt.UsageError, err)
	}
	defer func() { _ = f.Close() }()
	if g.DryRun {
		return g.Writer().Render(map[string]any{"method": "POST", "path": "/us_verifications", "file": c.Path})
	}
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	out := map[string]any{}
	form := url.Values{}
	if c.Description != "" {
		form.Set("description", c.Description)
	}
	resp, err := cl.Do(g.Context(), &client.Request{
		Method: "POST",
		Path:   "/us_verifications",
		Form:   form,
		Files:  []client.FilePart{{Field: "file", Filename: c.Path, Reader: f}},
		Out:    &out,
	})
	if err != nil {
		return err
	}
	if resp.Replayed {
		g.Writer().Notice("idempotent replay")
	}
	return g.Writer().Render(out)
}

// BulkCSVStatusCmd implements GET /v1/us_verifications/:id.
type BulkCSVStatusCmd struct {
	ID string `arg:"" help:"Job ID (us_csv_…)."`
}

// Run sends the request.
func (c *BulkCSVStatusCmd) Run(g *Globals) error {
	path, err := resourcePath("us_verifications", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// BulkCSVDownloadCmd implements GET /v1/us_verifications/:id/download.
type BulkCSVDownloadCmd struct {
	ID string `arg:"" help:"Job ID (us_csv_…)."`
}

// Run sends the request.
func (c *BulkCSVDownloadCmd) Run(g *Globals) error {
	path, err := resourcePath("us_verifications", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path+"/download", &out)
}

// BulkCSVListCmd implements GET /v1/us_verifications.
type BulkCSVListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *BulkCSVListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/us_verifications", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// BulkCSVDeleteCmd implements DELETE /v1/us_verifications/:id.
type BulkCSVDeleteCmd struct {
	ID      string `arg:"" help:"Job ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *BulkCSVDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("us_verifications", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execDelete(g, path, &out)
}
