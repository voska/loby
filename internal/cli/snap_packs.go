package cli

import (
	"net/url"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/lob"
)

// SnapPacksCmd implements /v1/snap_packs.
type SnapPacksCmd struct {
	Create SnapPackCreateCmd `cmd:"" help:"Send a snap pack."`
	Get    SnapPackGetCmd    `cmd:"" help:"Retrieve a snap pack by ID."`
	List   SnapPackListCmd   `cmd:"" help:"List snap packs."`
	Cancel SnapPackCancelCmd `cmd:"" help:"Cancel a snap pack before mailing."`
}

// SnapPackCreateCmd posts to /v1/snap_packs.
type SnapPackCreateCmd struct {
	Description    string            `help:"Internal description."`
	To             string            `help:"Recipient address ID or JSON." required:""`
	From           string            `help:"Sender address ID or JSON."`
	Outside        string            `help:"Outside artwork (HTML/URL/template/@file)." required:""`
	Inside         string            `help:"Inside artwork (HTML/URL/template/@file)." required:""`
	Size           string            `help:"Snap pack size." enum:"8.5x11" default:"8.5x11"`
	Color          bool              `help:"Print in color." default:"true"`
	MailingDate    string            `help:"Scheduled mailing date." name:"mailing-date"`
	MailType       string            `help:"Delivery class." enum:"usps_first_class" default:"usps_first_class" name:"mail-type"`
	UseType        string            `help:"Use type." enum:"marketing,operational" default:"operational" name:"use-type"`
	MergeVariables string            `help:"JSON object of template variables."`
	Metadata       map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *SnapPackCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description": optString(c.Description),
		"to":          parseAddressArg(c.To),
		"from":        parseAddressArg(c.From),
		"outside":     parseContentArg(c.Outside),
		"inside":      parseContentArg(c.Inside),
		"size":        c.Size,
		"color":       c.Color,
		"mail_type":   c.MailType,
		"use_type":    c.UseType,
		"metadata":    nilIfEmpty(c.Metadata),
	}
	if c.MailingDate != "" {
		body["mailing_date"] = c.MailingDate
	}
	if c.MergeVariables != "" {
		mv, err := parseJSONArg(c.MergeVariables)
		if err != nil {
			return errfmt.Wrap(errfmt.UsageError, err)
		}
		body["merge_variables"] = mv
	}
	pruneEmpty(body)
	return execCreateWithQuery(g, "snap_packs", "/snap_packs", url.Values{}, body, &lob.SnapPack{})
}

// SnapPackGetCmd implements GET /v1/snap_packs/:id.
type SnapPackGetCmd struct {
	ID string `arg:"" help:"Snap pack ID."`
}

// Run sends the request.
func (c *SnapPackGetCmd) Run(g *Globals) error {
	path, err := resourcePath("snap_packs", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.SnapPack{})
}

// SnapPackListCmd implements GET /v1/snap_packs.
type SnapPackListCmd struct {
	Limit        int    `help:"Max results (1-100)." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *SnapPackListCmd) Run(g *Globals) error {
	out := &lob.List[lob.SnapPack]{}
	return execList(g, "/snap_packs", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// SnapPackCancelCmd implements POST /v1/snap_packs/:id/cancel.
type SnapPackCancelCmd struct {
	ID      string `arg:"" help:"Snap pack ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *SnapPackCancelCmd) Run(g *Globals) error {
	return execCancel(g, "snap_packs", c.ID, c.Confirm, c.Force)
}
