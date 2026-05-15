package cli

import (
	"net/url"

	"github.com/voska/loby/internal/lob"
)

// BookletsCmd implements /v1/booklets.
type BookletsCmd struct {
	Create BookletCreateCmd `cmd:"" help:"Create a booklet."`
	Get    BookletGetCmd    `cmd:"" help:"Retrieve a booklet."`
	List   BookletListCmd   `cmd:"" help:"List booklets."`
	Delete BookletDeleteCmd `cmd:"" help:"Delete a booklet."`
}

// BookletCreateCmd posts to /v1/booklets.
type BookletCreateCmd struct {
	Description string            `help:"Internal description."`
	Cover       string            `help:"Cover artwork (HTML/URL/template/@file)." required:""`
	Inside      string            `help:"Inside artwork (multi-page PDF/HTML, or @file)." required:""`
	Size        string            `help:"Booklet size." enum:"8.5x11" default:"8.5x11"`
	Metadata    map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *BookletCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description": optString(c.Description),
		"cover":       parseContentArg(c.Cover),
		"inside":      parseContentArg(c.Inside),
		"size":        c.Size,
		"metadata":    nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	return execCreateWithQuery(g, "booklets", "/booklets", url.Values{}, body, &lob.Booklet{})
}

// BookletGetCmd implements GET /v1/booklets/:id.
type BookletGetCmd struct {
	ID string `arg:"" help:"Booklet ID."`
}

// Run sends the request.
func (c *BookletGetCmd) Run(g *Globals) error {
	path, err := resourcePath("booklets", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.Booklet{})
}

// BookletListCmd implements GET /v1/booklets.
type BookletListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *BookletListCmd) Run(g *Globals) error {
	out := &lob.List[lob.Booklet]{}
	return execList(g, "/booklets", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// BookletDeleteCmd implements DELETE /v1/booklets/:id.
type BookletDeleteCmd struct {
	ID      string `arg:"" help:"Booklet ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *BookletDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("booklets", c.ID)
	if err != nil {
		return err
	}
	return execDelete(g, path, &lob.Deleted{})
}
