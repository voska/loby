package cli

import (
	"errors"
	"net/url"

	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/lob"
)

// VerifyCmd is the top-level verification namespace. Mirrors `loby addresses verify`
// for discoverability but accepts country routing via subcommands.
type VerifyCmd struct {
	US   VerifyUSCmd   `cmd:"" help:"Verify a US address."`
	Intl VerifyIntlCmd `cmd:"" help:"Verify an international address."`
}

// VerifyUSCmd is the longer form of `loby addresses verify`.
type VerifyUSCmd struct {
	USVerifyCmd
}

// Run delegates to the addresses verify implementation.
func (c *VerifyUSCmd) Run(g *Globals) error {
	return c.USVerifyCmd.Run(g)
}

// VerifyIntlCmd verifies an international address.
type VerifyIntlCmd struct {
	Address   []string `arg:"" optional:"" help:"Single-line address."`
	Recipient string   `help:"Recipient name."`
	Primary   string   `help:"Primary line."`
	Secondary string   `help:"Secondary line."`
	City      string   `help:"City."`
	State     string   `help:"State / province / region."`
	Postal    string   `help:"Postal code."`
	Country   string   `help:"Two-letter ISO country code." required:""`
}

// Run sends the request.
func (c *VerifyIntlCmd) Run(g *Globals) error {
	body := lob.IntlVerificationCreate{
		Recipient:     c.Recipient,
		PrimaryLine:   c.Primary,
		SecondaryLine: c.Secondary,
		City:          c.City,
		State:         c.State,
		PostalCode:    c.Postal,
		Country:       c.Country,
	}
	if len(c.Address) > 0 {
		body.Address = joinSpace(c.Address)
	}
	if body.PrimaryLine == "" && body.Address == "" {
		return errfmt.Wrap(errfmt.UsageError, errors.New("provide a single-line address as positional, or --primary"))
	}
	return execCreateWithQuery(g, "intl_verifications", "/intl_verifications", url.Values{}, body, &lob.IntlVerification{})
}

// ZipCmd implements GET /v1/us_zip_lookups/:zip.
type ZipCmd struct {
	Zip string `arg:"" help:"5-digit US ZIP code."`
}

// Run sends the request.
func (c *ZipCmd) Run(g *Globals) error {
	return execGet(g, "/us_zip_lookups/"+c.Zip, &lob.ZipLookup{})
}
