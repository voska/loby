package cli

import "net/url"

// CreativesCmd implements /v1/creatives. Lob exposes only POST — creatives
// are write-once at create-time; the campaign drives delivery.
type CreativesCmd struct {
	Create CreativeCreateCmd `cmd:"" help:"Create a creative for a campaign."`
}

// CreativeCreateCmd posts to /v1/creatives. The shape Lob accepts is
// resource-type dependent (postcards take front/back, letters take file
// + extra_service, etc.) so the flags expose the union and the request
// builder only sends what's set.
type CreativeCreateCmd struct {
	CampaignID   string `help:"Parent campaign ID (cmp_…)." required:"" name:"campaign-id"`
	Description  string `help:"Internal description."`
	ResourceType string `help:"Mail piece kind." enum:"postcard,letter,self_mailer,snap_pack,booklet,card,buckslip" required:"" name:"resource-type"`
	From         string `help:"Sender address ID, inline JSON, or @file.json."`

	Front   string `help:"Front artwork: HTML string, URL, tmpl_id, or @file (postcard, self_mailer, snap_pack, buckslip)."`
	Back    string `help:"Back artwork: HTML string, URL, tmpl_id, or @file (postcard, self_mailer, snap_pack, buckslip)."`
	Inside  string `help:"Inside artwork: HTML string, URL, tmpl_id, or @file (self_mailer, snap_pack)."`
	Outside string `help:"Outside artwork: HTML string, URL, tmpl_id, or @file (self_mailer, snap_pack)."`
	Cover   string `help:"Cover artwork (booklet, card)."`
	File    string `help:"File body for letters (HTML, URL, tmpl_id, or @file)."`

	Size             string `help:"Mail piece size (e.g. 4x6, 6x11, 6x18 — varies by resource_type)."`
	MailType         string `help:"Postage class (usps_first_class, usps_standard, usps_standard_class)." name:"mail-type"`
	Color            bool   `help:"Color print for letters (defaults true on Lob's side)."`
	DoubleSided      bool   `help:"Letters double-sided." name:"double-sided"`
	AddressPlacement string `help:"Letter address placement (top_first_page, insert_blank_page)." name:"address-placement"`
	ExtraService     string `help:"Extra letter service (certified, registered, certified_return_receipt)." name:"extra-service"`

	Details  string            `help:"Override the details object (JSON or @file.json) — escape hatch for fields not exposed as flags."`
	Metadata map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *CreativeCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"campaign_id":   c.CampaignID,
		"resource_type": c.ResourceType,
		"description":   optString(c.Description),
		"metadata":      nilIfEmpty(c.Metadata),
	}
	if c.From != "" {
		body["from"] = parseAddressArg(c.From)
	}
	for _, p := range []struct{ flag, key, val string }{
		{"front", "front", c.Front},
		{"back", "back", c.Back},
		{"inside", "inside", c.Inside},
		{"outside", "outside", c.Outside},
		{"cover", "cover", c.Cover},
		{"file", "file", c.File},
	} {
		if p.val != "" {
			body[p.key] = parseContentArg(p.val)
		}
	}

	details := map[string]any{}
	if c.Size != "" {
		details["size"] = c.Size
	}
	if c.MailType != "" {
		details["mail_type"] = c.MailType
	}
	if c.Color {
		details["color"] = true
	}
	if c.DoubleSided {
		details["double_sided"] = true
	}
	if c.AddressPlacement != "" {
		details["address_placement"] = c.AddressPlacement
	}
	if c.ExtraService != "" {
		details["extra_service"] = c.ExtraService
	}
	// Explicit --details JSON overrides/extends the flag-derived map.
	if c.Details != "" {
		v, err := parseJSONArg(c.Details)
		if err != nil {
			return errfmtUsage("--details: " + err.Error())
		}
		if m, ok := v.(map[string]any); ok {
			for k, vv := range m {
				details[k] = vv
			}
		}
	}
	if len(details) > 0 {
		body["details"] = details
	}

	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "creatives", "/creatives", url.Values{}, body, &out)
}
