package cli

import (
	"net/url"
	"os"

	"github.com/voska/loby/internal/client"
	"github.com/voska/loby/internal/errfmt"
)

// UploadsCmd implements /v1/uploads — CSV file uploads for campaigns plus the
// exports/report sub-resources that surface row-level processing results.
type UploadsCmd struct {
	Create  UploadCreateCmd  `cmd:"" help:"Create an upload metadata record for a campaign."`
	File    UploadFileCmd    `cmd:"" help:"Upload a CSV file to an existing upload record."`
	Get     UploadGetCmd     `cmd:"" help:"Retrieve an upload (status is in the response body)."`
	List    UploadListCmd    `cmd:"" help:"List uploads."`
	Delete  UploadDeleteCmd  `cmd:"" help:"Delete an upload."`
	Exports UploadExportsCmd `cmd:"" help:"Manage upload export jobs (failed-row reports)."`
	Report  UploadReportCmd  `cmd:"" help:"Retrieve the line-item report for an upload (feature-flagged)."`
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

// UploadFileCmd posts a CSV file to POST /v1/uploads/:id/file (multipart).
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

// UploadGetCmd implements GET /v1/uploads/:id. The response includes the
// processing status as a field; agents should poll Get rather than relying on
// a separate /status endpoint (Lob does not expose one).
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

// UploadListCmd implements GET /v1/uploads. The endpoint only accepts a
// campaignId filter — Lob rejects pagination params with HTTP 400.
type UploadListCmd struct {
	CampaignID string `help:"Filter by campaign ID." name:"campaign-id"`
}

// Run sends the request. Note: /uploads returns a bare JSON array, not Lob's
// usual {data:[…]} envelope — hence the []map[string]any rather than List.
func (c *UploadListCmd) Run(g *Globals) error {
	q := url.Values{}
	if c.CampaignID != "" {
		q.Set("campaignId", c.CampaignID)
	}
	var out []map[string]any
	return execList(g, "/uploads", q, &out)
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

// UploadExportsCmd manages export jobs that produce row-by-row error reports.
type UploadExportsCmd struct {
	Create UploadExportCreateCmd `cmd:"" help:"Create an export job for an upload."`
	List   UploadExportListCmd   `cmd:"" help:"List export jobs for an upload."`
	Get    UploadExportGetCmd    `cmd:"" help:"Retrieve a specific export job."`
}

// UploadExportCreateCmd implements POST /v1/uploads/:id/exports.
type UploadExportCreateCmd struct {
	ID   string `arg:"" help:"Upload ID."`
	Type string `help:"Export type." enum:"failures,all" default:"failures"`
}

// Run sends the request.
func (c *UploadExportCreateCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	body := map[string]any{"type": c.Type}
	out := map[string]any{}
	return execCreateWithQuery(g, "uploads.exports", path+"/exports", url.Values{}, body, &out)
}

// UploadExportListCmd implements GET /v1/uploads/:id/exports.
type UploadExportListCmd struct {
	ID string `arg:"" help:"Upload ID."`
}

// Run sends the request.
func (c *UploadExportListCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path+"/exports", &out)
}

// UploadExportGetCmd implements GET /v1/uploads/:id/exports/:ex_id.
type UploadExportGetCmd struct {
	ID       string `arg:"" help:"Upload ID."`
	ExportID string `arg:"" help:"Export job ID."`
}

// Run sends the request.
func (c *UploadExportGetCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	export, err := resourcePath("exports", c.ExportID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path+export, &out)
}

// UploadReportCmd implements GET /v1/uploads/:id/report. The endpoint is
// feature-flagged by Lob — agents may see 404 until enabled on their account.
type UploadReportCmd struct {
	ID     string `arg:"" help:"Upload ID."`
	Status string `help:"Filter by line-item status." enum:"Validated,Failed,Processing,${none}" default:"${none}"`
	Limit  int    `help:"Max rows (1-100)." default:"100"`
	Offset int    `help:"Pagination offset."`
}

// Run sends the request.
func (c *UploadReportCmd) Run(g *Globals) error {
	path, err := resourcePath("uploads", c.ID)
	if err != nil {
		return err
	}
	q := url.Values{}
	if c.Status != "" {
		q.Set("status", c.Status)
	}
	if c.Limit > 0 {
		q.Set("limit", itoa(c.Limit))
	}
	if c.Offset > 0 {
		q.Set("offset", itoa(c.Offset))
	}
	out := map[string]any{}
	encoded := q.Encode()
	if encoded != "" {
		encoded = "?" + encoded
	}
	return execGet(g, path+"/report"+encoded, &out)
}
