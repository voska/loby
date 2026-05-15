package cli

import (
	"net/url"

	"github.com/voska/loby/internal/lob"
)

// CardsCmd implements /v1/cards (printed card stock for campaigns).
type CardsCmd struct {
	Create CardCreateCmd `cmd:"" help:"Create a card."`
	Get    CardGetCmd    `cmd:"" help:"Retrieve a card."`
	List   CardListCmd   `cmd:"" help:"List cards."`
	Delete CardDeleteCmd `cmd:"" help:"Delete a card."`
}

// CardCreateCmd posts to /v1/cards.
type CardCreateCmd struct {
	Description     string            `help:"Internal description."`
	Front           string            `help:"Front artwork (HTML/URL/template/@file)." required:""`
	Back            string            `help:"Back artwork (HTML/URL/template/@file)."`
	Size            string            `help:"Card size." enum:"2.125x3.375,2.5x2.5,2.75x2.75,2x3.5,3.5x2,3.5x4,3.5x5,4x6,5x5,5x7,6x9,8.5x11" default:"3.5x2"`
	Stock           string            `help:"Stock type." enum:"14PT_AQ,14PT_AQ_DULL,14PT_UV,14PT_MATTE,16PT_AQ,16PT_UV,${none}" default:"${none}"`
	AutoReorder     bool              `help:"Auto-reorder when low." name:"auto-reorder"`
	ReorderQuantity int               `help:"Quantity per reorder." name:"reorder-quantity"`
	Metadata        map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *CardCreateCmd) Run(g *Globals) error {
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
	return execCreateWithQuery(g, "cards", "/cards", url.Values{}, body, &lob.Card{})
}

// CardGetCmd / List / Delete

// CardGetCmd implements GET /v1/cards/:id.
type CardGetCmd struct {
	ID string `arg:"" help:"Card ID (card_…)."`
}

// Run sends the request.
func (c *CardGetCmd) Run(g *Globals) error {
	path, err := resourcePath("cards", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.Card{})
}

// CardListCmd implements GET /v1/cards.
type CardListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *CardListCmd) Run(g *Globals) error {
	out := &lob.List[lob.Card]{}
	return execList(g, "/cards", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// CardDeleteCmd implements DELETE /v1/cards/:id.
type CardDeleteCmd struct {
	ID      string `arg:"" help:"Card ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *CardDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("cards", c.ID)
	if err != nil {
		return err
	}
	return execDelete(g, path, &lob.Deleted{})
}
