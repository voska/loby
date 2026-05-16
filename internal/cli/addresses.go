package cli

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/voska/loby/internal/client"
	"github.com/voska/loby/internal/errfmt"
	"github.com/voska/loby/internal/lob"
)

// AddressesCmd groups address CRUD plus the verify/autocomplete/zip helpers.
type AddressesCmd struct {
	Create       AddressCreateCmd  `cmd:"" help:"Save an address to the Lob address book."`
	Get          AddressGetCmd     `cmd:"" help:"Retrieve a saved address by ID."`
	List         AddressListCmd    `cmd:"" help:"List saved addresses."`
	Delete       AddressDeleteCmd  `cmd:"" help:"Delete a saved address."`
	Verify       USVerifyCmd       `cmd:"" help:"Verify a US address (shortcut for 'verify us')."`
	Autocomplete USAutocompleteCmd `cmd:"" help:"Autocomplete a US address prefix."`
}

// AddressCreateCmd implements POST /v1/addresses.
type AddressCreateCmd struct {
	Description string            `help:"Internal description (≤255 chars)."`
	Name        string            `help:"Recipient name (required if Company is blank)."`
	Company     string            `help:"Recipient company (required if Name is blank)."`
	Email       string            `help:"Recipient email."`
	Phone       string            `help:"Recipient phone."`
	Line1       string            `help:"Street line 1." required:"" name:"line1"`
	Line2       string            `help:"Street line 2." name:"line2"`
	City        string            `help:"City."`
	State       string            `help:"State (two-letter US code or full name)."`
	Zip         string            `help:"ZIP code or international postal code."`
	Country     string            `help:"Two-letter ISO country code." default:"US"`
	Metadata    map[string]string `help:"Metadata key=value pairs (repeatable)."`
}

// Run sends the request.
func (c *AddressCreateCmd) Run(g *Globals) error {
	if c.Name == "" && c.Company == "" {
		return errfmt.Wrap(errfmt.UsageError, errors.New("either --name or --company is required"))
	}
	body := lob.AddressCreate{
		Description:    c.Description,
		Name:           c.Name,
		Company:        c.Company,
		Email:          c.Email,
		Phone:          c.Phone,
		AddressLine1:   c.Line1,
		AddressLine2:   c.Line2,
		AddressCity:    c.City,
		AddressState:   c.State,
		AddressZip:     c.Zip,
		AddressCountry: c.Country,
		Metadata:       c.Metadata,
	}
	return execCreate(g, "addresses", "/addresses", body, &lob.Address{})
}

// AddressGetCmd implements GET /v1/addresses/:id.
type AddressGetCmd struct {
	ID string `arg:"" help:"Address ID (adr_…)."`
}

// Run sends the request.
func (c *AddressGetCmd) Run(g *Globals) error {
	path, err := resourcePath("addresses", c.ID)
	if err != nil {
		return err
	}
	return execGet(g, path, &lob.Address{})
}

// AddressListCmd implements GET /v1/addresses.
type AddressListCmd struct {
	Limit        int    `help:"Max results per page (1-100)." default:"10"`
	Before       string `help:"Pagination cursor: items before this ID."`
	After        string `help:"Pagination cursor: items after this ID."`
	IncludeTotal bool   `help:"Include total count in response." name:"include-total"`
}

// Run sends the request and emits NDJSON when listing more than the limit.
func (c *AddressListCmd) Run(g *Globals) error {
	q := url.Values{}
	if c.Limit > 0 {
		q.Set("limit", strconv.Itoa(c.Limit))
	}
	if c.Before != "" {
		q.Set("before", c.Before)
	}
	if c.After != "" {
		q.Set("after", c.After)
	}
	if c.IncludeTotal {
		q.Set("include[]", "total_count")
	}
	out := &lob.List[lob.Address]{}
	return execList(g, "/addresses", q, out)
}

// AddressDeleteCmd implements DELETE /v1/addresses/:id.
type AddressDeleteCmd struct {
	ID      string `arg:"" help:"Address ID (adr_…)."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *AddressDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("addresses", c.ID)
	if err != nil {
		return err
	}
	return execDelete(g, path, &lob.Deleted{})
}

// USVerifyCmd verifies a US address. Accepts either a single-line address
// positional or structured flags.
type USVerifyCmd struct {
	Address   []string `arg:"" optional:"" help:"Single-line address (e.g. \"185 Berry St, San Francisco, CA 94107\")."`
	Recipient string   `help:"Recipient name."`
	Primary   string   `help:"Primary line (street)."`
	Secondary string   `help:"Secondary line (apt/suite)."`
	City      string   `help:"City."`
	State     string   `help:"State."`
	Zip       string   `help:"ZIP code."`
	Case      string   `help:"Case transformation." enum:"upper,proper,default" default:"default"`
}

// Run sends the request.
func (c *USVerifyCmd) Run(g *Globals) error {
	body := lob.USVerificationCreate{
		Recipient:     c.Recipient,
		PrimaryLine:   c.Primary,
		SecondaryLine: c.Secondary,
		City:          c.City,
		State:         c.State,
		ZipCode:       c.Zip,
	}
	if len(c.Address) > 0 {
		body.Address = joinSpace(c.Address)
	}
	if body.PrimaryLine == "" && body.Address == "" {
		return errfmt.Wrap(errfmt.UsageError, errors.New("provide a single-line address as positional, or --primary"))
	}
	q := url.Values{}
	if c.Case != "" && c.Case != "default" {
		q.Set("case", c.Case)
	}
	return execCreateWithQuery(g, "us_verifications", "/us_verifications", q, body, &lob.USVerification{})
}

// USAutocompleteCmd suggests completions for a partial US address.
type USAutocompleteCmd struct {
	Prefix    []string `arg:"" help:"Partial street address prefix."`
	City      string   `help:"Optional city filter."`
	State     string   `help:"Optional state filter."`
	Zip       string   `help:"Optional ZIP filter."`
	GeoIPSort bool     `help:"Bias suggestions by client IP."`
	Case      string   `help:"Case transformation." enum:"upper,proper,default" default:"default"`
}

// Run sends the request.
func (c *USAutocompleteCmd) Run(g *Globals) error {
	if len(c.Prefix) == 0 {
		return errfmt.Wrap(errfmt.UsageError, errors.New("address prefix required"))
	}
	body := lob.USAutocompletionCreate{
		AddressPrefix: joinSpace(c.Prefix),
		City:          c.City,
		State:         c.State,
		ZipCode:       c.Zip,
		GeoIPSort:     c.GeoIPSort,
	}
	q := url.Values{}
	if c.Case != "" && c.Case != "default" {
		q.Set("case", c.Case)
	}
	return execCreateWithQuery(g, "us_autocompletions", "/us_autocompletions", q, body, &lob.USAutocompletion{})
}

func joinSpace(xs []string) string {
	out := ""
	for i, s := range xs {
		if i > 0 {
			out += " "
		}
		out += s
	}
	return out
}

// execCreate is the shared POST + JSON body helper.
func execCreate(g *Globals, command, path string, body, out any) error {
	return execCreateWithQuery(g, command, path, nil, body, out)
}

func execCreateWithQuery(g *Globals, command, path string, q url.Values, body, out any) error {
	if g.DryRun {
		return g.Writer().Render(map[string]any{
			"method": http.MethodPost,
			"path":   path,
			"query":  q,
			"body":   body,
		})
	}
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	idem := g.IdempotencyKey
	if idem == "" {
		k, err := client.IdempotencyKey(command, nil, body, true)
		if err != nil {
			return err
		}
		idem = k
	}
	resp, err := cl.Do(g.Context(), &client.Request{
		Method:         http.MethodPost,
		Path:           path,
		Query:          q,
		Body:           body,
		Out:            out,
		IdempotencyKey: idem,
	})
	if err != nil {
		return err
	}
	if resp.Replayed {
		g.Writer().Notice("idempotent replay (Lob returned cached response)")
	}
	return g.Writer().Render(out)
}

func execGet(g *Globals, path string, out any) error {
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	if _, err := cl.Do(g.Context(), &client.Request{
		Method: http.MethodGet,
		Path:   path,
		Out:    out,
	}); err != nil {
		return err
	}
	return g.Writer().Render(out)
}

func execList(g *Globals, path string, q url.Values, out any) error {
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	if _, err := cl.Do(g.Context(), &client.Request{
		Method: http.MethodGet,
		Path:   path,
		Query:  q,
		Out:    out,
	}); err != nil {
		return err
	}
	return g.Writer().Render(out)
}

func execDelete(g *Globals, path string, out any) error {
	if g.DryRun {
		return g.Writer().Render(map[string]any{
			"method": http.MethodDelete,
			"path":   path,
		})
	}
	cl, err := g.LobClient()
	if err != nil {
		return err
	}
	if _, err := cl.Do(g.Context(), &client.Request{
		Method: http.MethodDelete,
		Path:   path,
		Out:    out,
	}); err != nil {
		return err
	}
	return g.Writer().Render(out)
}
