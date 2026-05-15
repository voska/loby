package cli

import (
	"net/url"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/lob"
)

// LettersCmd implements /v1/letters.
type LettersCmd struct {
	Create LetterCreateCmd `cmd:"" help:"Send a letter."`
	Get    LetterGetCmd    `cmd:"" help:"Retrieve a letter by ID."`
	List   LetterListCmd   `cmd:"" help:"List letters."`
	Cancel LetterCancelCmd `cmd:"" help:"Cancel a letter before mailing."`
}

// LetterCreateCmd posts to /v1/letters.
type LetterCreateCmd struct {
	Description      string            `help:"Internal description."`
	To               string            `help:"Recipient address ID or JSON." required:""`
	From             string            `help:"Sender address ID or JSON." required:""`
	File             string            `help:"PDF, HTML, URL, or template ID. Use @file.pdf for a local file." required:""`
	Color            bool              `help:"Print in color."`
	DoubleSided      bool              `help:"Double-sided printing." name:"double-sided" default:"true"`
	AddressPlacement string            `help:"Address placement." enum:"top_first_page,insert_blank_page" default:"top_first_page" name:"address-placement"`
	MailingDate      string            `help:"Scheduled mailing date." name:"mailing-date"`
	MailType         string            `help:"Delivery class." enum:"usps_first_class,usps_standard" default:"usps_first_class" name:"mail-type"`
	UseType          string            `help:"Use type." enum:"marketing,operational" default:"operational" name:"use-type"`
	ExtraService     string            `help:"Tracked / certified service." enum:"certified,registered,certified_return_receipt,${none}" default:"${none}" name:"extra-service"`
	PerforatedPage   int               `help:"Page index to perforate (1-based)." name:"perforated-page"`
	ReturnEnvelope   string            `help:"Return envelope ID."`
	CustomEnvelope   string            `help:"Custom envelope ID."`
	MergeVariables   string            `help:"JSON object of template variables (or @file.json)."`
	Metadata         map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *LetterCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description":       optString(c.Description),
		"to":                parseAddressArg(c.To),
		"from":              parseAddressArg(c.From),
		"file":              parseContentArg(c.File),
		"color":             c.Color,
		"double_sided":      c.DoubleSided,
		"address_placement": c.AddressPlacement,
		"mail_type":         c.MailType,
		"use_type":          c.UseType,
		"metadata":          nilIfEmpty(c.Metadata),
	}
	if c.MailingDate != "" {
		body["mailing_date"] = c.MailingDate
	}
	if c.ExtraService != "" {
		body["extra_service"] = c.ExtraService
	}
	if c.PerforatedPage > 0 {
		body["perforated_page"] = c.PerforatedPage
	}
	if c.ReturnEnvelope != "" {
		body["return_envelope"] = c.ReturnEnvelope
	}
	if c.CustomEnvelope != "" {
		body["custom_envelope"] = c.CustomEnvelope
	}
	if c.MergeVariables != "" {
		mv, err := parseJSONArg(c.MergeVariables)
		if err != nil {
			return errfmt.Wrap(errfmt.UsageError, err)
		}
		body["merge_variables"] = mv
	}
	pruneEmpty(body)
	return execCreateWithQuery(g, "letters", "/letters", url.Values{}, body, &lob.Letter{})
}

// LetterGetCmd implements GET /v1/letters/:id.
type LetterGetCmd struct {
	ID string `arg:"" help:"Letter ID (ltr_…)."`
}

// Run sends the request.
func (c *LetterGetCmd) Run(g *Globals) error {
	path, err := resourcePath("letters", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.Letter{})
}

// LetterListCmd implements GET /v1/letters.
type LetterListCmd struct {
	Limit        int    `help:"Max results (1-100)." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *LetterListCmd) Run(g *Globals) error {
	out := &lob.List[lob.Letter]{}
	return execList(g, "/letters", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// LetterCancelCmd implements POST /v1/letters/:id/cancel.
type LetterCancelCmd struct {
	ID      string `arg:"" help:"Letter ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *LetterCancelCmd) Run(g *Globals) error {
	return execCancel(g, "letters", c.ID, c.Confirm, c.Force)
}
