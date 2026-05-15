package cli

import "net/url"

// BillingGroupsCmd implements /v1/billing_groups.
type BillingGroupsCmd struct {
	Create BillingGroupCreateCmd `cmd:"" help:"Create a billing group."`
	Get    BillingGroupGetCmd    `cmd:"" help:"Retrieve a billing group."`
	List   BillingGroupListCmd   `cmd:"" help:"List billing groups."`
	Update BillingGroupUpdateCmd `cmd:"" help:"Update a billing group."`
}

// BillingGroupCreateCmd posts to /v1/billing_groups.
type BillingGroupCreateCmd struct {
	Name        string            `help:"Group name." required:""`
	Description string            `help:"Group description."`
	Metadata    map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *BillingGroupCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"name":        c.Name,
		"description": optString(c.Description),
		"metadata":    nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "billing_groups", "/billing_groups", url.Values{}, body, &out)
}

// BillingGroupGetCmd implements GET /v1/billing_groups/:id.
type BillingGroupGetCmd struct {
	ID string `arg:"" help:"Billing group ID."`
}

// Run sends the request.
func (c *BillingGroupGetCmd) Run(g *Globals) error {
	path, err := resourcePath("billing_groups", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// BillingGroupListCmd implements GET /v1/billing_groups.
type BillingGroupListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *BillingGroupListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/billing_groups", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// BillingGroupUpdateCmd implements POST /v1/billing_groups/:id.
type BillingGroupUpdateCmd struct {
	ID          string            `arg:"" help:"Billing group ID."`
	Name        string            `help:"New name."`
	Description string            `help:"New description."`
	Metadata    map[string]string `help:"Replace metadata."`
}

// Run sends the request.
func (c *BillingGroupUpdateCmd) Run(g *Globals) error {
	path, err := resourcePath("billing_groups", c.ID)
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
	if len(c.Metadata) > 0 {
		body["metadata"] = c.Metadata
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "billing_groups.update", path, url.Values{}, body, &out)
}
