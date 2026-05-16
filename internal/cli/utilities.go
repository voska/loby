package cli

import "net/url"

// QRCodesCmd implements /v1/qr_code_analytics — Lob's QR code analytics
// endpoint. QR codes themselves are created by embedding Lob's QR snippet in a
// mailer's HTML; the API only surfaces scan analytics for the resulting codes.
type QRCodesCmd struct {
	List QRCodeListCmd `cmd:"" help:"List QR codes (with scan analytics)."`
}

// QRCodeListCmd implements GET /v1/qr_code_analytics.
type QRCodeListCmd struct {
	Limit        int  `help:"Max results." default:"10"`
	Offset       int  `help:"Pagination offset."`
	Scanned      bool `help:"Only QR codes with at least one scan event."`
	IncludeTotal bool `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *QRCodeListCmd) Run(g *Globals) error {
	q := url.Values{}
	if c.Limit > 0 {
		q.Set("limit", itoa(c.Limit))
	}
	if c.Offset > 0 {
		q.Set("offset", itoa(c.Offset))
	}
	if c.Scanned {
		q.Set("scanned", "true")
	}
	if c.IncludeTotal {
		q.Set("include[]", "total_count")
	}
	out := map[string]any{}
	return execList(g, "/qr_code_analytics", q, &out)
}

// LinksCmd implements /v1/links — Lob's URL shortener. Links are short URLs
// (optionally rooted on a custom domain, see [DomainsCmd]) that redirect to a
// long URL and track clicks.
type LinksCmd struct {
	Create LinkCreateCmd `cmd:"" help:"Create a short link."`
	Get    LinkGetCmd    `cmd:"" help:"Retrieve a short link."`
	List   LinkListCmd   `cmd:"" help:"List short links."`
	Delete LinkDeleteCmd `cmd:"" help:"Delete a short link."`
}

// LinkCreateCmd posts to /v1/links.
type LinkCreateCmd struct {
	RedirectURL string            `help:"Long URL the short link redirects to." required:"" name:"redirect-link"`
	DomainID    string            `help:"Optional custom domain ID (defaults to Lob's short domain)." name:"domain-id"`
	Description string            `help:"Internal description."`
	Metadata    map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *LinkCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"redirect_link": c.RedirectURL,
		"domain_id":     optString(c.DomainID),
		"description":   optString(c.Description),
		"metadata":      nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "links", "/links", url.Values{}, body, &out)
}

// LinkGetCmd implements GET /v1/links/:id.
type LinkGetCmd struct {
	ID string `arg:"" help:"Link ID."`
}

// Run sends the request.
func (c *LinkGetCmd) Run(g *Globals) error {
	path, err := resourcePath("links", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// LinkListCmd implements GET /v1/links.
type LinkListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *LinkListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/links", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// LinkDeleteCmd implements DELETE /v1/links/:id.
type LinkDeleteCmd struct {
	ID      string `arg:"" help:"Link ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *LinkDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("links", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execDelete(g, path, &out)
}

// DomainsCmd implements /v1/domains — custom short-link domains. Use the
// returned domain ID with `loby links create --domain-id` to root your short
// URLs on your own domain instead of Lob's.
type DomainsCmd struct {
	Create DomainCreateCmd `cmd:"" help:"Register a custom domain for use with the URL shortener."`
	Get    DomainGetCmd    `cmd:"" help:"Retrieve a domain."`
	List   DomainListCmd   `cmd:"" help:"List domains."`
	Delete DomainDeleteCmd `cmd:"" help:"Delete a domain."`
}

// DomainCreateCmd posts to /v1/domains.
type DomainCreateCmd struct {
	Domain      string            `help:"Domain (e.g. links.example.com)." required:""`
	Description string            `help:"Internal description."`
	Metadata    map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *DomainCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"domain":      c.Domain,
		"description": optString(c.Description),
		"metadata":    nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "domains", "/domains", url.Values{}, body, &out)
}

// DomainGetCmd implements GET /v1/domains/:id.
type DomainGetCmd struct {
	ID string `arg:"" help:"Domain ID."`
}

// Run sends the request.
func (c *DomainGetCmd) Run(g *Globals) error {
	path, err := resourcePath("domains", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// DomainListCmd implements GET /v1/domains.
type DomainListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *DomainListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/domains", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// DomainDeleteCmd implements DELETE /v1/domains/:id.
type DomainDeleteCmd struct {
	ID      string `arg:"" help:"Domain ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *DomainDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("domains", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execDelete(g, path, &out)
}

// GeoCmd implements /v1/reverse_geocode_lookups.
type GeoCmd struct {
	Reverse GeoReverseCmd `cmd:"" help:"Reverse-geocode lat/lng to ZIP codes."`
}

// GeoReverseCmd posts to /v1/reverse_geocode_lookups. Coordinates use flags
// (not positionals) so negative values like --lng=-122.4194 don't get
// interpreted as flag short-names by the parser.
type GeoReverseCmd struct {
	Latitude  float64 `help:"Latitude (e.g. 37.7749)." required:"" name:"lat"`
	Longitude float64 `help:"Longitude (e.g. -122.4194)." required:"" name:"lng"`
}

// Run sends the request.
func (c *GeoReverseCmd) Run(g *Globals) error {
	body := map[string]any{"latitude": c.Latitude, "longitude": c.Longitude}
	out := map[string]any{}
	return execCreateWithQuery(g, "reverse_geocode_lookups", "/us_reverse_geocode_lookups", url.Values{}, body, &out)
}

// IdentityCmd implements /v1/identity_validation.
type IdentityCmd struct {
	Verify IdentityVerifyCmd `cmd:"" help:"Verify the identity of a recipient."`
	Get    IdentityGetCmd    `cmd:"" help:"Retrieve an identity validation."`
}

// IdentityVerifyCmd posts to /v1/identity_validation.
type IdentityVerifyCmd struct {
	FirstName    string `help:"First name." required:"" name:"first-name"`
	LastName     string `help:"Last name." required:"" name:"last-name"`
	AddressLine1 string `help:"Address line 1." required:"" name:"line1"`
	AddressLine2 string `help:"Address line 2." name:"line2"`
	City         string `help:"City."`
	State        string `help:"State."`
	Zip          string `help:"ZIP code."`
	Country      string `help:"Country code." default:"US"`
}

// Run sends the request.
func (c *IdentityVerifyCmd) Run(g *Globals) error {
	body := map[string]any{
		"first_name":      c.FirstName,
		"last_name":       c.LastName,
		"address_line1":   c.AddressLine1,
		"address_line2":   optString(c.AddressLine2),
		"address_city":    optString(c.City),
		"address_state":   optString(c.State),
		"address_zip":     optString(c.Zip),
		"address_country": c.Country,
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "identity_validation", "/identity_validation", url.Values{}, body, &out)
}

// IdentityGetCmd implements GET /v1/identity_validation/:id.
type IdentityGetCmd struct {
	ID string `arg:"" help:"Identity validation ID."`
}

// Run sends the request.
func (c *IdentityGetCmd) Run(g *Globals) error {
	path, err := resourcePath("identity_validation", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// ResourceProofsCmd implements /v1/resource_proofs.
type ResourceProofsCmd struct {
	Get ResourceProofGetCmd `cmd:"" help:"Retrieve a resource proof (PDF preview of a printed asset)."`
}

// ResourceProofGetCmd implements GET /v1/resource_proofs/:id.
type ResourceProofGetCmd struct {
	ID string `arg:"" help:"Resource proof ID."`
}

// Run sends the request.
func (c *ResourceProofGetCmd) Run(g *Globals) error {
	path, err := resourcePath("resource_proofs", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}
