package cli

import (
	"net/url"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/lob"
)

// SelfMailersCmd implements /v1/self_mailers.
type SelfMailersCmd struct {
	Create SelfMailerCreateCmd `cmd:"" help:"Send a self-mailer."`
	Get    SelfMailerGetCmd    `cmd:"" help:"Retrieve a self-mailer by ID."`
	List   SelfMailerListCmd   `cmd:"" help:"List self-mailers."`
	Cancel SelfMailerCancelCmd `cmd:"" help:"Cancel a self-mailer before mailing."`
}

// SelfMailerCreateCmd posts to /v1/self_mailers.
type SelfMailerCreateCmd struct {
	Description    string            `help:"Internal description."`
	To             string            `help:"Recipient address ID or JSON." required:""`
	From           string            `help:"Sender address ID or JSON."`
	Outside        string            `help:"Outside artwork: HTML/URL/template ID/@file." required:""`
	Inside         string            `help:"Inside artwork: HTML/URL/template ID/@file." required:""`
	Size           string            `help:"Self-mailer size." enum:"6x18_bifold,12x9_bifold,11x9_bifold" default:"6x18_bifold"`
	MailingDate    string            `help:"Scheduled mailing date." name:"mailing-date"`
	MailType       string            `help:"Delivery class." enum:"usps_first_class,usps_standard" default:"usps_first_class" name:"mail-type"`
	UseType        string            `help:"Use type." enum:"marketing,operational" default:"marketing" name:"use-type"`
	MergeVariables string            `help:"JSON object of template variables (or @file.json)."`
	Metadata       map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *SelfMailerCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description": optString(c.Description),
		"to":          parseAddressArg(c.To),
		"from":        parseAddressArg(c.From),
		"outside":     parseContentArg(c.Outside),
		"inside":      parseContentArg(c.Inside),
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
	return execCreateWithQuery(g, "self_mailers", "/self_mailers", url.Values{}, body, &lob.SelfMailer{})
}

// SelfMailerGetCmd implements GET /v1/self_mailers/:id.
type SelfMailerGetCmd struct {
	ID string `arg:"" help:"Self-mailer ID (sfm_…)."`
}

// Run sends the request.
func (c *SelfMailerGetCmd) Run(g *Globals) error {
	path, err := resourcePath("self_mailers", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.SelfMailer{})
}

// SelfMailerListCmd implements GET /v1/self_mailers.
type SelfMailerListCmd struct {
	Limit        int    `help:"Max results (1-100)." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *SelfMailerListCmd) Run(g *Globals) error {
	out := &lob.List[lob.SelfMailer]{}
	return execList(g, "/self_mailers", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// SelfMailerCancelCmd implements POST /v1/self_mailers/:id/cancel.
type SelfMailerCancelCmd struct {
	ID      string `arg:"" help:"Self-mailer ID (sfm_…)."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *SelfMailerCancelCmd) Run(g *Globals) error {
	return execCancel(g, "self_mailers", c.ID, c.Confirm, c.Force)
}
