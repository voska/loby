package cli

import (
	"net/url"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/lob"
)

// PostcardsCmd implements /v1/postcards. Lob does not expose a cancel/delete
// endpoint for postcards — they enter the USPS pipeline immediately on create.
type PostcardsCmd struct {
	Create PostcardCreateCmd `cmd:"" help:"Send a postcard."`
	Get    PostcardGetCmd    `cmd:"" help:"Retrieve a postcard by ID."`
	List   PostcardListCmd   `cmd:"" help:"List postcards."`
}

// PostcardCreateCmd is the full create form. To/From accept address IDs (adr_…)
// or inline JSON objects via --to-json / --from-json.
type PostcardCreateCmd struct {
	Description    string            `help:"Internal description (≤255 chars)."`
	To             string            `help:"Recipient address ID (adr_…) or JSON object." required:""`
	From           string            `help:"Sender address ID (adr_…) or JSON object."`
	Front          string            `help:"HTML, URL, template ID (tmpl_…), or @file.html for the front." required:""`
	Back           string            `help:"HTML, URL, template ID (tmpl_…), or @file.html for the back."`
	Size           string            `help:"Postcard size." enum:"4x6,6x9,6x11" default:"4x6"`
	MailingDate    string            `help:"Scheduled mailing date (YYYY-MM-DD or RFC3339)." name:"mailing-date"`
	MailType       string            `help:"Delivery class." enum:"usps_first_class,usps_standard" default:"usps_first_class" name:"mail-type"`
	UseType        string            `help:"Use type." enum:"marketing,operational" default:"marketing" name:"use-type"`
	MergeVariables string            `help:"JSON object of template variables (or @file.json)."`
	Metadata       map[string]string `help:"Metadata key=value pairs (repeatable)."`
}

// Run builds the request body and posts to /postcards.
func (c *PostcardCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description": optString(c.Description),
		"to":          parseAddressArg(c.To),
		"from":        parseAddressArg(c.From),
		"front":       parseContentArg(c.Front),
		"back":        parseContentArg(c.Back),
		"size":        c.Size,
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
	return execCreateWithQuery(g, "postcards", "/postcards", url.Values{}, body, &lob.Postcard{})
}

// PostcardGetCmd implements GET /v1/postcards/:id.
type PostcardGetCmd struct {
	ID string `arg:"" help:"Postcard ID (psc_…)."`
}

// Run sends the request.
func (c *PostcardGetCmd) Run(g *Globals) error {
	path, err := resourcePath("postcards", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.Postcard{})
}

// PostcardListCmd implements GET /v1/postcards.
type PostcardListCmd struct {
	Limit        int    `help:"Max results (1-100)." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *PostcardListCmd) Run(g *Globals) error {
	out := &lob.List[lob.Postcard]{}
	return execList(g, "/postcards", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// Helpers shared across mailers.

func optString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nilIfEmpty(m map[string]string) any {
	if len(m) == 0 {
		return nil
	}
	return m
}

func pruneEmpty(m map[string]any) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		}
	}
}
