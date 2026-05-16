package cli

import "net/url"

// CampaignsCmd implements /v1/campaigns.
type CampaignsCmd struct {
	Create CampaignCreateCmd `cmd:"" help:"Create a campaign."`
	Get    CampaignGetCmd    `cmd:"" help:"Retrieve a campaign."`
	List   CampaignListCmd   `cmd:"" help:"List campaigns."`
	Update CampaignUpdateCmd `cmd:"" help:"Update a campaign."`
	Delete CampaignDeleteCmd `cmd:"" help:"Delete a campaign."`
	Send   CampaignSendCmd   `cmd:"" help:"Submit a campaign for processing (no longer editable)."`
}

// CampaignCreateCmd posts to /v1/campaigns.
type CampaignCreateCmd struct {
	Name         string            `help:"Campaign name." required:""`
	Description  string            `help:"Internal description."`
	ScheduleType string            `help:"Schedule type." enum:"immediate,in_future" default:"immediate" name:"schedule-type"`
	SendDate     string            `help:"Send date (RFC3339, required for schedule-type=in_future)." name:"send-date"`
	BillingGroup string            `help:"Billing group ID." name:"billing-group-id"`
	Metadata     map[string]string `help:"Metadata key=value pairs."`
	UseType      string            `help:"Use type." enum:"marketing,operational" default:"marketing" name:"use-type"`
}

// Run sends the request.
func (c *CampaignCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"name":          c.Name,
		"description":   optString(c.Description),
		"schedule_type": c.ScheduleType,
		"use_type":      c.UseType,
		"metadata":      nilIfEmpty(c.Metadata),
	}
	if c.SendDate != "" {
		body["send_date"] = c.SendDate
	}
	if c.BillingGroup != "" {
		body["billing_group_id"] = c.BillingGroup
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "campaigns", "/campaigns", url.Values{}, body, &out)
}

// CampaignGetCmd implements GET /v1/campaigns/:id.
type CampaignGetCmd struct {
	ID string `arg:"" help:"Campaign ID (cmp_…)."`
}

// Run sends the request.
func (c *CampaignGetCmd) Run(g *Globals) error {
	path, err := resourcePath("campaigns", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// CampaignListCmd implements GET /v1/campaigns.
type CampaignListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *CampaignListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/campaigns", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// CampaignUpdateCmd implements POST /v1/campaigns/:id.
type CampaignUpdateCmd struct {
	ID           string            `arg:"" help:"Campaign ID."`
	Name         string            `help:"New name."`
	Description  string            `help:"New description."`
	SendDate     string            `help:"New send date." name:"send-date"`
	BillingGroup string            `help:"Billing group ID." name:"billing-group-id"`
	Metadata     map[string]string `help:"Replace metadata."`
}

// Run sends the request.
func (c *CampaignUpdateCmd) Run(g *Globals) error {
	path, err := resourcePath("campaigns", c.ID)
	if err != nil {
		return err
	}
	body := map[string]any{}
	if c.Name != "" {
		body["name"] = c.Name
	}
	if c.Description != "" {
		body["description"] = c.Description
	}
	if c.SendDate != "" {
		body["send_date"] = c.SendDate
	}
	if c.BillingGroup != "" {
		body["billing_group_id"] = c.BillingGroup
	}
	if len(c.Metadata) > 0 {
		body["metadata"] = c.Metadata
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "campaigns.update", path, url.Values{}, body, &out)
}

// CampaignDeleteCmd implements DELETE /v1/campaigns/:id.
type CampaignDeleteCmd struct {
	ID      string `arg:"" help:"Campaign ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *CampaignDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("campaigns", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execDelete(g, path, &out)
}

// CampaignSendCmd implements POST /v1/campaigns/:id/send.
type CampaignSendCmd struct {
	ID      string `arg:"" help:"Campaign ID."`
	Confirm bool   `help:"Required to send (campaign cannot be edited after)." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *CampaignSendCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("campaigns", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "campaigns.send", path+"/send", url.Values{}, map[string]any{}, &out)
}

// InformedDeliveryCmd implements /v1/informed_delivery_campaigns. Per Lob's
// public spec, create takes multipart/form-data with a JPG ride-along image.
type InformedDeliveryCmd struct {
	Create InformedDeliveryCreateCmd `cmd:"" help:"Create an informed delivery campaign."`
	Get    InformedDeliveryGetCmd    `cmd:"" help:"Retrieve an informed delivery campaign."`
	List   InformedDeliveryListCmd   `cmd:"" help:"List informed delivery campaigns."`
}

// InformedDeliveryCreateCmd posts a multipart/form-data request to
// /v1/informed_delivery_campaigns. Lob requires ride_along_url + ride_along_image
// (JPG, ≤200KB, ≤300x200) + quantity + start_date.
type InformedDeliveryCreateCmd struct {
	RideAlongURL        string `help:"Click-through URL the ride-along image links to (https://, ≤255 chars)." required:"" name:"ride-along-url"`
	RideAlongImage      string `help:"Path to ride-along JPG (@file.jpg or absolute path; ≤200KB, ≤300x200)." required:"" name:"ride-along-image"`
	Quantity            int    `help:"Number of mail pieces (USPS IMB sequence)." required:""`
	StartDate           string `help:"First hand-off date (YYYY-MM-DD, ≥2 days from now)." required:"" name:"start-date"`
	Status              string `help:"Initial status." enum:"approved,pending_approval" default:"approved"`
	BrandName           string `help:"Brand name for the campaign." name:"brand-name"`
	RepresentativeImage string `help:"Optional representative JPG (same constraints as ride-along)." name:"representative-image"`
}

// Run sends the multipart request.
func (c *InformedDeliveryCreateCmd) Run(g *Globals) error {
	form := url.Values{}
	form.Set("ride_along_url", c.RideAlongURL)
	form.Set("quantity", itoa(c.Quantity))
	form.Set("start_date", c.StartDate)
	if c.Status != "" {
		form.Set("status", c.Status)
	}
	if c.BrandName != "" {
		form.Set("brand_name", c.BrandName)
	}

	files, cleanup, err := openImageParts(map[string]string{
		"ride_along_image":     c.RideAlongImage,
		"representative_image": c.RepresentativeImage,
	})
	if err != nil {
		return err
	}
	defer cleanup()

	return execMultipart(g, "informed_delivery_campaigns", "/informed_delivery_campaigns", form, files, map[string]any{})
}

// InformedDeliveryGetCmd implements GET /v1/informed_delivery_campaigns/:id.
type InformedDeliveryGetCmd struct {
	ID string `arg:"" help:"Informed delivery campaign ID."`
}

// Run sends the request.
func (c *InformedDeliveryGetCmd) Run(g *Globals) error {
	path, err := resourcePath("informed_delivery_campaigns", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// InformedDeliveryListCmd implements GET /v1/informed_delivery_campaigns.
type InformedDeliveryListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *InformedDeliveryListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/informed_delivery_campaigns", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}
