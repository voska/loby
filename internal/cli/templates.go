package cli

import "net/url"

// TemplatesCmd implements /v1/templates.
type TemplatesCmd struct {
	Create   TemplateCreateCmd   `cmd:"" help:"Create a template (HTML stored at Lob with merge variables)."`
	Get      TemplateGetCmd      `cmd:"" help:"Retrieve a template."`
	List     TemplateListCmd     `cmd:"" help:"List templates."`
	Update   TemplateUpdateCmd   `cmd:"" help:"Update a template's description or metadata."`
	Delete   TemplateDeleteCmd   `cmd:"" help:"Delete a template."`
	Versions TemplateVersionsCmd `cmd:"" help:"Manage template versions."`
}

// TemplateCreateCmd posts to /v1/templates.
type TemplateCreateCmd struct {
	Description string            `help:"Internal description."`
	HTML        string            `help:"HTML body (or @file.html)." required:"" name:"html"`
	EngineType  string            `help:"Template engine." enum:"legacy,handlebars" default:"handlebars" name:"engine-type"`
	Metadata    map[string]string `help:"Metadata key=value pairs."`
}

// Run sends the request.
func (c *TemplateCreateCmd) Run(g *Globals) error {
	body := map[string]any{
		"description": optString(c.Description),
		"html":        parseContentArg(c.HTML),
		"engine_type": c.EngineType,
		"metadata":    nilIfEmpty(c.Metadata),
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "templates", "/templates", url.Values{}, body, &out)
}

// TemplateGetCmd implements GET /v1/templates/:id.
type TemplateGetCmd struct {
	ID string `arg:"" help:"Template ID (tmpl_…)."`
}

// Run sends the request.
func (c *TemplateGetCmd) Run(g *Globals) error {
	path, err := resourcePath("templates", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// TemplateListCmd implements GET /v1/templates.
type TemplateListCmd struct {
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *TemplateListCmd) Run(g *Globals) error {
	out := map[string]any{}
	return execList(g, "/templates", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// TemplateUpdateCmd implements POST /v1/templates/:id.
type TemplateUpdateCmd struct {
	ID               string            `arg:"" help:"Template ID."`
	Description      string            `help:"New description."`
	PublishedVersion string            `help:"ID of the version to mark as published." name:"published-version"`
	Metadata         map[string]string `help:"Replace metadata."`
}

// Run sends the request.
func (c *TemplateUpdateCmd) Run(g *Globals) error {
	path, err := resourcePath("templates", c.ID)
	if err != nil {
		return err
	}
	body := map[string]any{}
	if c.Description != "" {
		body["description"] = c.Description
	}
	if c.PublishedVersion != "" {
		body["published_version"] = c.PublishedVersion
	}
	if len(c.Metadata) > 0 {
		body["metadata"] = c.Metadata
	}
	out := map[string]any{}
	return execCreateWithQuery(g, "templates.update", path, url.Values{}, body, &out)
}

// TemplateDeleteCmd implements DELETE /v1/templates/:id.
type TemplateDeleteCmd struct {
	ID      string `arg:"" help:"Template ID."`
	Confirm bool   `help:"Required for destructive operations." xor:"destructive"`
	Force   bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *TemplateDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("templates", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execDelete(g, path, &out)
}

// TemplateVersionsCmd groups template-version operations.
type TemplateVersionsCmd struct {
	Create TemplateVersionCreateCmd `cmd:"" help:"Create a new template version (publishes HTML)."`
	Get    TemplateVersionGetCmd    `cmd:"" help:"Retrieve a template version."`
	List   TemplateVersionListCmd   `cmd:"" help:"List template versions."`
	Update TemplateVersionUpdateCmd `cmd:"" help:"Update a template version's description."`
	Delete TemplateVersionDeleteCmd `cmd:"" help:"Delete a template version."`
}

// TemplateVersionCreateCmd posts to /v1/templates/:tmpl_id/versions.
type TemplateVersionCreateCmd struct {
	TemplateID  string `arg:"" help:"Parent template ID (tmpl_…)."`
	Description string `help:"Version description."`
	HTML        string `help:"HTML body (or @file.html)." required:""`
	EngineType  string `help:"Template engine." enum:"legacy,handlebars" default:"handlebars" name:"engine-type"`
}

// Run sends the request.
func (c *TemplateVersionCreateCmd) Run(g *Globals) error {
	path, err := resourcePath("templates", c.TemplateID)
	if err != nil {
		return err
	}
	body := map[string]any{
		"description": optString(c.Description),
		"html":        parseContentArg(c.HTML),
		"engine_type": c.EngineType,
	}
	pruneEmpty(body)
	out := map[string]any{}
	return execCreateWithQuery(g, "template_versions", path+"/versions", url.Values{}, body, &out)
}

// TemplateVersionGetCmd implements GET /v1/templates/:tmpl_id/versions/:id.
type TemplateVersionGetCmd struct {
	TemplateID string `arg:"" help:"Parent template ID."`
	VersionID  string `arg:"" help:"Version ID (vrsn_…)."`
}

// Run sends the request.
func (c *TemplateVersionGetCmd) Run(g *Globals) error {
	path, err := resourcePath("templates", c.TemplateID)
	if err != nil {
		return err
	}
	if invalidID := invalidVersion(c.VersionID); invalidID != nil {
		return invalidID
	}
	out := map[string]any{}
	return execGet(g, path+"/versions/"+c.VersionID, &out)
}

// TemplateVersionListCmd implements GET /v1/templates/:tmpl_id/versions.
type TemplateVersionListCmd struct {
	TemplateID   string `arg:"" help:"Parent template ID."`
	Limit        int    `help:"Max results." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	IncludeTotal bool   `help:"Include total count." name:"include-total"`
}

// Run sends the request.
func (c *TemplateVersionListCmd) Run(g *Globals) error {
	path, err := resourcePath("templates", c.TemplateID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execList(g, path+"/versions", listQuery(c.Limit, c.Before, c.After, c.IncludeTotal, nil), &out)
}

// TemplateVersionUpdateCmd implements POST /v1/templates/:tmpl_id/versions/:id.
type TemplateVersionUpdateCmd struct {
	TemplateID  string `arg:"" help:"Parent template ID."`
	VersionID   string `arg:"" help:"Version ID."`
	Description string `help:"New description." required:""`
}

// Run sends the request.
func (c *TemplateVersionUpdateCmd) Run(g *Globals) error {
	path, err := resourcePath("templates", c.TemplateID)
	if err != nil {
		return err
	}
	if invalidID := invalidVersion(c.VersionID); invalidID != nil {
		return invalidID
	}
	body := map[string]any{"description": c.Description}
	out := map[string]any{}
	return execCreateWithQuery(g, "template_versions.update", path+"/versions/"+c.VersionID, url.Values{}, body, &out)
}

// TemplateVersionDeleteCmd implements DELETE /v1/templates/:tmpl_id/versions/:id.
type TemplateVersionDeleteCmd struct {
	TemplateID string `arg:"" help:"Parent template ID."`
	VersionID  string `arg:"" help:"Version ID."`
	Confirm    bool   `help:"Required for destructive operations." xor:"destructive"`
	Force      bool   `help:"Alias for --confirm." xor:"destructive"`
}

// Run sends the request.
func (c *TemplateVersionDeleteCmd) Run(g *Globals) error {
	if err := requireConfirm(c.Confirm, c.Force); err != nil {
		return err
	}
	path, err := resourcePath("templates", c.TemplateID)
	if err != nil {
		return err
	}
	if invalidID := invalidVersion(c.VersionID); invalidID != nil {
		return invalidID
	}
	out := map[string]any{}
	return execDelete(g, path+"/versions/"+c.VersionID, &out)
}

func invalidVersion(id string) error {
	for _, r := range id {
		if r < 0x20 || r == '/' || r == '?' || r == '#' || r == '%' {
			return resourcePathErr(id)
		}
	}
	if id == "" {
		return resourcePathErr("")
	}
	return nil
}

func resourcePathErr(id string) error {
	_, err := resourcePath("", id)
	if err == nil {
		return nil
	}
	return err
}
