package cli

import "net/url"

// QRCodesCmd implements /v1/qr_codes.
type QRCodesCmd struct {
	Create QRCodeCreateCmd `cmd:"" help:"Create a QR code with a redirect URL."`
	Get    QRCodeGetCmd    `cmd:"" help:"Retrieve a QR code."`
	List   QRCodeListCmd   `cmd:"" help:"List QR codes."`
}

// QRCodeCreateCmd posts to /v1/qr_codes.
type QRCodeCreateCmd struct {
	RedirectURL string            `help:"Destination URL the code resolves to." required:"" name:"redirect-url"`
	Description string            `help:"Internal description."`
	Position    string            `help:"Position on artwork." enum:"top_left,top_right,bottom_left,bottom_right,relative" default:"bottom_right"`
	Width       string            `help:"Width in inches."`
	Top         string            `help:"Top offset in inches."`
	Right       string            `help:"Right offset in inches."`
	Bottom      string            `help:"Bottom offset in inches."`
	Left        string            `help:"Left offset in inches."`
	Pages       string            `help:"Pages to place QR code on (e.g. 'front', 'back', '1-3')."`
	Metadata    map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *QRCodeCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"redirect_url": c.RedirectURL,
		"description":  optString(c.Description),
		"position":     c.Position,
		"width":        optString(c.Width),
		"top":          optString(c.Top),
		"right":        optString(c.Right),
		"bottom":       optString(c.Bottom),
		"left":         optString(c.Left),
		"pages":        optString(c.Pages),
		"metadata":     nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "qr_codes", "/qr_codes", url.Values{}, body, &out)
}

// QRCodeGetCmd implements GET /v1/qr_codes/:id.
type QRCodeGetCmd struct {
	ID string `arg:"" help:"QR code ID."`
}

// Run sends the request.
func (c *QRCodeGetCmd) Run(g *Globals) error {
	path, err := resourcePath("qr_codes", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// QRCodeListCmd implements GET /v1/qr_codes.
type QRCodeListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *QRCodeListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/qr_codes", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// URLShortenerCmd implements /v1/short_urls.
type URLShortenerCmd struct {
	Create URLShortenerCreateCmd `cmd:"" help:"Create a tracked short URL."`
	Get    URLShortenerGetCmd    `cmd:"" help:"Retrieve a short URL."`
	List   URLShortenerListCmd   `cmd:"" help:"List short URLs."`
}

// URLShortenerCreateCmd posts to /v1/short_urls.
type URLShortenerCreateCmd struct {
	RedirectURL string            `help:"Destination URL." required:"" name:"redirect-url"`
	Description string            `help:"Internal description."`
	Metadata    map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *URLShortenerCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"redirect_url": c.RedirectURL,
		"description":  optString(c.Description),
		"metadata":     nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "short_urls", "/short_urls", url.Values{}, body, &out)
}

// URLShortenerGetCmd implements GET /v1/short_urls/:id.
type URLShortenerGetCmd struct {
	ID string `arg:"" help:"Short URL ID."`
}

// Run sends the request.
func (c *URLShortenerGetCmd) Run(g *Globals) error {
	path, err := resourcePath("short_urls", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// URLShortenerListCmd implements GET /v1/short_urls.
type URLShortenerListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *URLShortenerListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/short_urls", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// GeoCmd implements /v1/reverse_geocode_lookups.
type GeoCmd struct {
	Reverse GeoReverseCmd `cmd:"" help:"Reverse-geocode lat/lng to ZIP codes."`
}

// GeoReverseCmd posts to /v1/reverse_geocode_lookups.
type GeoReverseCmd struct {
	Latitude  float64 `arg:"" help:"Latitude."`
	Longitude float64 `arg:"" help:"Longitude."`
}

// Run sends the request.
func (c *GeoReverseCmd) Run(g *Globals) error {
	body := map[string]any{"latitude": c.Latitude, "longitude": c.Longitude}
	out := map[string]any{}
	return execCreateWithQuery(g, "reverse_geocode_lookups", "/reverse_geocode_lookups", url.Values{}, body, &out)
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
