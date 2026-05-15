package cli

import (
	"net/url"
	"os"

	"github.com/voska/loby/internal/client"
	"github.com/voska/loby/internal/errfmt"
)

// UploadsCmd implements /v1/uploads — CSV file uploads for campaigns.
type UploadsCmd struct {
	Create UploadCreateCmd `cmd:"" help:"Create an upload metadata record for a campaign."`
	File   UploadFileCmd   `cmd:"" help:"Upload a CSV file to an existing upload record."`
	Get    UploadGetCmd    `cmd:"" help:"Retrieve an upload."`
	List   UploadListCmd   `cmd:"" help:"List uploads."`
	Delete UploadDeleteCmd `cmd:"" help:"Delete an upload."`
	Status UploadStatusCmd `cmd:"" help:"Get upload processing status."`
	Errors UploadErrorsCmd `cmd:"" help:"Download upload validation errors as CSV."`
}

// UploadCreateCmd posts to /v1/uploads.
type UploadCreateCmd struct {
	CampaignID            string            `help:"Campaign ID (cmp_…)." required:"" name:"campaign-id"`
	ColumnMapping         string            `help:"Column mapping JSON (or @file.json)." name:"column-mapping"`
	Metadata              map[string]string `help:"Metadata key=value pairs."`
	RequiredAddressColumn string            `help:"Required address column name." name:"required-address-column"`
}

// Run sends the request.
func (c *UploadCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"campaignId": c.CampaignID,
		"metadata":   nilIfEmpty(c.Metadata),
	}
	if c.ColumnMapping != "" {
		mv, err := parseJSONArg(c.ColumnMapping)
		if err != nil {
			return errfmt.Wrap(errfmt.UsageError, err)
		}
		body["columnMapping"] = mv
	}
	if c.RequiredAddressColumn != "" {
		body["requiredAddressColumn"] = c.RequiredAddressColumn
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "uploads", "/uploads", url.Values{}, body, &out)
}

// UploadFileCmd posts a CSV file to /v1/uploads/:id/file (multipart).
type UploadFileCmd struct {
	ID   string `arg:"" help:"Upload ID."`
	Path string `arg:"" help:"Path to CSV file."`
}

// Run sends the multipart request.
func (c *UploadFileCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	f, err := os.Open(c.Path) //nolint:gosec // path is a user-supplied CLI argument
	if err != nil {
		return errfmt.Wrap(errfmt.UsageError, err)
	}
	defer func() { _ = f.Close() }()
	if g.DryRun {
		return g.Writer().Render(map[string]any{"method": "POST", "path": path + "/file", "file": c.Path})
	}
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	out := map[string]any{}
	resp, err := cl.Do(g.Context(), &client.Request{
		Method: "POST",
		Path:   path + "/file",
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

// UploadGetCmd implements GET /v1/uploads/:id.
type UploadGetCmd struct {
	ID string `arg:"" help:"Upload ID."`
}

// Run sends the request.
func (c *UploadGetCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// UploadListCmd implements GET /v1/uploads.
type UploadListCmd struct {
	CampaignID string `help:"Filter by campaign ID." name:"campaign-id"`
	Limit      int    `help:"Max results." default:"10"`
	Before     string `help:"Pagination cursor before."`
	After      string `help:"Pagination cursor after."`
}

// Run sends the request.
func (c *UploadListCmd) Run(g *Globals) error {
	extra := url.Values{}
	if c.CampaignID != "" {
		extra.Set("campaignId", c.CampaignID)
	}
	out := map[string]any{}
	return execList(g, "/uploads", listQuery(c.Limit, c.Before, c.After, false, extra), &out)
}

// UploadDeleteCmd implements DELETE /v1/uploads/:id.
type UploadDeleteCmd struct {
	ID      string `arg:"" help:"Upload ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *UploadDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execDelete(g, path, &out)
}

// UploadStatusCmd implements GET /v1/uploads/:id/status.
type UploadStatusCmd struct {
	ID string `arg:"" help:"Upload ID."`
}

// Run sends the request.
func (c *UploadStatusCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path+"/status", &out)
}

// UploadErrorsCmd implements GET /v1/uploads/:id/rows/errors.
type UploadErrorsCmd struct {
	ID string `arg:"" help:"Upload ID."`
}

// Run sends the request.
func (c *UploadErrorsCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path+"/rows/errors", &out)
}
