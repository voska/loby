package cli

import (
	"net/url"

	"github.com/voska/loby/internal/errfmt"
)

// BulkCmd implements Lob's two documented bulk-verification endpoints. Lob does
// not currently expose a public async-CSV verification API — the historical
// `us_csv_verifications` resource is unlisted in `lob-api-public.yml`. If/when
// Lob re-publishes it, add a `csv` subcommand here.
type BulkCmd struct {
	US   BulkUSCmd   `cmd:"" help:"Bulk US address verifications (synchronous, ≤100 addresses)."`
	Intl BulkIntlCmd `cmd:"" help:"Bulk international address verifications (synchronous, ≤100 addresses)."`
}

// BulkUSCmd posts up to 100 addresses to POST /bulk/us_verifications.
type BulkUSCmd struct {
	Addresses string `help:"JSON array of address objects (or @file.json)." required:""`
	Case      string `help:"Case transformation." enum:"upper,proper,default" default:"default"`
}

// Run sends the request.
func (c *BulkUSCmd) Run(g *Globals) error {
	body, err := parseJSONArg(c.Addresses)
	if err != nil {
		return errfmt.Wrap(errfmt.UsageError, err)
	}
	q := url.Values{}
	if c.Case != "" && c.Case != "default" {
		q.Set("case", c.Case)
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "bulk_us_verifications", "/bulk/us_verifications", q, map[string]any{"addresses": body}, &out)
}

// BulkIntlCmd posts up to 100 addresses to POST /bulk/intl_verifications.
type BulkIntlCmd struct {
	Addresses string `help:"JSON array of address objects (or @file.json)." required:""`
}

// Run sends the request.
func (c *BulkIntlCmd) Run(g *Globals) error {
	body, err := parseJSONArg(c.Addresses)
	if err != nil {
		return errfmt.Wrap(errfmt.UsageError, err)
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "bulk_intl_verifications", "/bulk/intl_verifications", url.Values{}, map[string]any{"addresses": body}, &out)
}
