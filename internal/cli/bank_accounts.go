package cli

import (
	"errors"
	"net/url"

	"github.com/voska/loby/internal/errfmt"
)

// BankAccountsCmd implements /v1/bank_accounts.
type BankAccountsCmd struct {
	Create BankAccountCreateCmd `cmd:"" help:"Add a bank account."`
	Get    BankAccountGetCmd    `cmd:"" help:"Retrieve a bank account."`
	List   BankAccountListCmd   `cmd:"" help:"List bank accounts."`
	Delete BankAccountDeleteCmd `cmd:"" help:"Delete a bank account."`
	Verify BankAccountVerifyCmd `cmd:"" help:"Verify a bank account with the test deposits."`
}

// BankAccountCreateCmd posts to /v1/bank_accounts.
type BankAccountCreateCmd struct {
	Description   string            `help:"Internal description."`
	RoutingNumber string            `help:"9-digit routing number." required:"" name:"routing-number"`
	AccountNumber string            `help:"Account number." required:"" name:"account-number"`
	AccountType   string            `help:"Account type." enum:"individual,company" required:"" name:"account-type"`
	Signatory     string            `help:"Full name of signatory." required:""`
	Metadata      map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *BankAccountCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description":    optString(c.Description),
		"routing_number": c.RoutingNumber,
		"account_number": c.AccountNumber,
		"account_type":   c.AccountType,
		"signatory":      c.Signatory,
		"metadata":       nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "bank_accounts", "/bank_accounts", url.Values{}, body, &out)
}

// BankAccountGetCmd implements GET /v1/bank_accounts/:id.
type BankAccountGetCmd struct {
	ID string `arg:"" help:"Bank account ID (bank_…)."`
}

// Run sends the request.
func (c *BankAccountGetCmd) Run(g *Globals) error {
	path, err := resourcePath("bank_accounts", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// BankAccountListCmd implements GET /v1/bank_accounts.
type BankAccountListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *BankAccountListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/bank_accounts", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// BankAccountDeleteCmd implements DELETE /v1/bank_accounts/:id.
type BankAccountDeleteCmd struct {
	ID      string `arg:"" help:"Bank account ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *BankAccountDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("bank_accounts", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execDelete(g, path, &out)
}

// BankAccountVerifyCmd implements POST /v1/bank_accounts/:id/verify.
type BankAccountVerifyCmd struct {
	ID      string `arg:"" help:"Bank account ID."`
	Amounts []int  `help:"Two test-deposit amounts in cents (e.g. --amounts 11,35)." required:""`
}

// Run sends the request.
func (c *BankAccountVerifyCmd) Run(g *Globals) error {
	if len(c.Amounts) != 2 {
		return errfmt.Wrap(errfmt.UsageError, errors.New("exactly two --amounts are required"))
	}
	path, err := resourcePath("bank_accounts", c.ID)
	if err != nil {
		return err
	}
	body := map[string]any{"amounts": c.Amounts}
	out := map[string]any{}
	return execCreateWithQuery(g, "bank_accounts.verify", path+"/verify", url.Values{}, body, &out)
}
