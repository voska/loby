package cli

import (
	"net/url"

	"github.com/voska/loby/internal/lob"
)

// BuckslipsCmd implements /v1/buckslips.
type BuckslipsCmd struct {
	Create BuckslipCreateCmd `cmd:"" help:"Create a buckslip."`
	Get    BuckslipGetCmd    `cmd:"" help:"Retrieve a buckslip."`
	List   BuckslipListCmd   `cmd:"" help:"List buckslips."`
	Delete BuckslipDeleteCmd `cmd:"" help:"Delete a buckslip."`
}

// BuckslipCreateCmd posts to /v1/buckslips.
type BuckslipCreateCmd struct {
	Description     string            `help:"Internal description."`
	Front           string            `help:"Front artwork (HTML/URL/template/@file)." required:""`
	Back            string            `help:"Back artwork (HTML/URL/template/@file)."`
	Size            string            `help:"Buckslip size." enum:"8.75x3.75" default:"8.75x3.75"`
	Stock           string            `help:"Stock type."`
	AutoReorder     bool              `help:"Auto-reorder when low." name:"auto-reorder"`
	ReorderQuantity int               `help:"Quantity per reorder." name:"reorder-quantity"`
	Metadata        map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *BuckslipCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description":  optString(c.Description),
		"front":        parseContentArg(c.Front),
		"back":         parseContentArg(c.Back),
		"size":         c.Size,
		"stock":        optString(c.Stock),
		"auto_reorder": c.AutoReorder,
		"metadata":     nilIfEmpty(c.Metadata),
	}
	if c.ReorderQuantity > 0 {
		body["reorder_quantity"] = c.ReorderQuantity
	}
	pruneEmpty(body)
	return execCreateWithQuery(g, "buckslips", "/buckslips", url.Values{}, body, &lob.Buckslip{})
}

// BuckslipGetCmd implements GET /v1/buckslips/:id.
type BuckslipGetCmd struct {
	ID string `arg:"" help:"Buckslip ID."`
}

// Run sends the request.
func (c *BuckslipGetCmd) Run(g *Globals) error {
	path, err := resourcePath("buckslips", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.Buckslip{})
}

// BuckslipListCmd implements GET /v1/buckslips.
type BuckslipListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *BuckslipListCmd) Run(g *Globals) error {
	out := &lob.List[lob.Buckslip]{}
	return execList(g, "/buckslips", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// BuckslipDeleteCmd implements DELETE /v1/buckslips/:id.
type BuckslipDeleteCmd struct {
	ID      string `arg:"" help:"Buckslip ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *BuckslipDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("buckslips", c.ID)
	if err != nil {
		return err
	}
	return execDelete(g, path, &lob.Deleted{})
}
