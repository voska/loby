package cli

import "net/url"

// CreativesCmd implements /v1/creatives — the artwork attached to a campaign.
type CreativesCmd struct {
	Create CreativeCreateCmd `cmd:"" help:"Create a creative for a campaign."`
	Get    CreativeGetCmd    `cmd:"" help:"Retrieve a creative."`
	Update CreativeUpdateCmd `cmd:"" help:"Update a creative."`
}

// CreativeCreateCmd posts to /v1/creatives.
type CreativeCreateCmd struct {
	CampaignID   string            `help:"Parent campaign ID (cmp_…)." required:"" name:"campaign-id"`
	Description  string            `help:"Internal description."`
	ResourceType string            `help:"Resource type." enum:"postcard,letter,self_mailer,snap_pack" required:"" name:"resource-type"`
	From         string            `help:"Sender address ID or JSON." required:""`
	Details      string            `help:"Detail JSON (HTML/file refs etc., or @file.json)."`
	MailType     string            `help:"Delivery class." enum:"usps_first_class,usps_standard" default:"usps_first_class" name:"mail-type"`
	Metadata     map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *CreativeCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"campaign_id":   c.CampaignID,
		"description":   optString(c.Description),
		"resource_type": c.ResourceType,
		"from":          parseAddressArg(c.From),
		"mail_type":     c.MailType,
		"metadata":      nilIfEmpty(c.Metadata),
	}
	if c.Details != "" {
		v, err := parseJSONArg(c.Details)
		if err == nil {
			if m, ok := v.(map[string]any); ok {
				for k, vv := range m {
					body[k] = vv
				}
			} else {
				body["details"] = v
			}
		}
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "creatives", "/creatives", url.Values{}, body, &out)
}

// CreativeGetCmd implements GET /v1/creatives/:id.
type CreativeGetCmd struct {
	ID string `arg:"" help:"Creative ID (crv_…)."`
}

// Run sends the request.
func (c *CreativeGetCmd) Run(g *Globals) error {
	path, err := resourcePath("creatives", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// CreativeUpdateCmd implements POST /v1/creatives/:id.
type CreativeUpdateCmd struct {
	ID          string            `arg:"" help:"Creative ID."`
	Description string            `help:"New description."`
	Metadata    map[string]string `help:"Replace metadata."`
}

// Run sends the request.
func (c *CreativeUpdateCmd) Run(g *Globals) error {
	path, err := resourcePath("creatives", c.ID)
	if err != nil {
		return err
	}
	body := map[string]any{}
	if c.Description != "" {
		body["description"] = c.Description
	}
	if len(c.Metadata) > 0 {
		body["metadata"] = c.Metadata
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "creatives.update", path, url.Values{}, body, &out)
}
