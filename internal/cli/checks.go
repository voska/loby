package cli

import (
	"errors"
	"net/url"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/lob"
)

// ChecksCmd implements /v1/checks.
type ChecksCmd struct {
	Create CheckCreateCmd `cmd:"" help:"Mail a check."`
	Get    CheckGetCmd    `cmd:"" help:"Retrieve a check by ID."`
	List   CheckListCmd   `cmd:"" help:"List checks."`
	Cancel CheckCancelCmd `cmd:"" help:"Cancel a check before mailing."`
}

// CheckCreateCmd posts to /v1/checks.
type CheckCreateCmd struct {
	Description    string            `help:"Internal description."`
	To             string            `help:"Recipient address ID or JSON." required:""`
	From           string            `help:"Sender address ID or JSON."`
	BankAccount    string            `help:"Bank account ID (bank_…)." required:"" name:"bank-account"`
	Amount         float64           `help:"Amount in dollars (max $9,999,999.99)." required:""`
	Memo           string            `help:"Memo line (≤40 chars)."`
	Message        string            `help:"Inset message printed below the check."`
	CheckNumber    int               `help:"Specific check number (default: next sequential)." name:"check-number"`
	Logo           string            `help:"Logo: URL, ID, or @file.png."`
	Attachment     string            `help:"Single-page attachment (PDF/HTML/URL/template, or @file)."`
	MailingDate    string            `help:"Scheduled mailing date." name:"mailing-date"`
	MailType       string            `help:"Delivery class." enum:"usps_first_class,usps_standard" default:"usps_first_class" name:"mail-type"`
	UseType        string            `help:"Use type." enum:"marketing,operational" default:"operational" name:"use-type"`
	MergeVariables string            `help:"JSON object of template variables (or @file.json)."`
	Metadata       map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *CheckCreateCmd) Run(g *Globals) error {
	if c.Amount <= 0 {
		return errfmt.Wrap(errfmt.UsageError, errors.New("--amount must be > 0"))
	}
	body := map[string]any{
		"description":  optString(c.Description),
		"to":           parseAddressArg(c.To),
		"from":         parseAddressArg(c.From),
		"bank_account": c.BankAccount,
		"amount":       c.Amount,
		"memo":         optString(c.Memo),
		"message":      optString(c.Message),
		"logo":         parseContentArg(c.Logo),
		"attachment":   parseContentArg(c.Attachment),
		"mail_type":    c.MailType,
		"use_type":     c.UseType,
		"metadata":     nilIfEmpty(c.Metadata),
	}
	if c.CheckNumber > 0 {
		body["check_number"] = c.CheckNumber
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
	return execCreateWithQuery(g, "checks", "/checks", url.Values{}, body, &lob.Check{})
}

// CheckGetCmd implements GET /v1/checks/:id.
type CheckGetCmd struct {
	ID string `arg:"" help:"Check ID (chk_…)."`
}

// Run sends the request.
func (c *CheckGetCmd) Run(g *Globals) error {
	path, err := resourcePath("checks", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.Check{})
}

// CheckListCmd implements GET /v1/checks.
type CheckListCmd struct {
	Limit        int    `help:"Max results (1-100)." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *CheckListCmd) Run(g *Globals) error {
	out := &lob.List[lob.Check]{}
	return execList(g, "/checks", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), out)
}

// CheckCancelCmd implements POST /v1/checks/:id/cancel.
type CheckCancelCmd struct {
	ID      string `arg:"" help:"Check ID (chk_…)."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *CheckCancelCmd) Run(g *Globals) error {
	return execCancel(g, "checks", c.ID, c.Confirm, c.Force)
}
