package cli

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// EventsCmd implements /v1/events.
type EventsCmd struct {
	List EventListCmd `cmd:"" help:"List events."`
	Tail EventTailCmd `cmd:"" help:"Stream events as NDJSON (poll every 5s)."`
	Get  EventGetCmd  `cmd:"" help:"Retrieve a single event by ID."`
}

// EventListCmd implements GET /v1/events.
type EventListCmd struct {
	Limit        int    `help:"Max results (1-100)." default:"10"`
	Before       string `help:"Pagination cursor before."`
	After        string `help:"Pagination cursor after."`
	EventType    string `help:"Filter by event type (e.g. postcard.created)." name:"event-type"`
	ResourceType string `help:"Filter by resource type." name:"resource-type"`
	DateCreated  string `help:"Filter on date_created (RFC3339 range, e.g. 'gte:2026-01-01,lt:2026-02-01')." name:"date-created"`
}

// Run sends the request.
func (c *EventListCmd) Run(g *Globals) error {
	extra := url.Values{}
	if c.EventType != "" {
		extra.Set("event_type", c.EventType)
	}
	if c.ResourceType != "" {
		extra.Set("resource_type", c.ResourceType)
	}
	if c.DateCreated != "" {
		extra.Set("date_created", c.DateCreated)
	}
	out := map[string]any{}
	return execList(g, "/events", listQuery(c.Limit, c.Before, c.After, false, extra), &out)
}

// EventGetCmd implements GET /v1/events/:id.
type EventGetCmd struct {
	ID string `arg:"" help:"Event ID (evt_…)."`
}

// Run sends the request.
func (c *EventGetCmd) Run(g *Globals) error {
	path, err := resourcePath("events", c.ID)
	if err != nil {
		return err
	}
	out := map[string]any{}
	return execGet(g, path, &out)
}

// EventTailCmd polls the event log and streams new events as NDJSON. Designed
// for agents that need a long-running tail without webhooks.
type EventTailCmd struct {
	Interval     time.Duration `help:"Poll interval." default:"5s"`
	ResourceType string        `help:"Filter by resource type." name:"resource-type"`
	EventType    string        `help:"Filter by event type." name:"event-type"`
}

// Run polls events until the context is canceled.
func (c *EventTailCmd) Run(g *Globals) error {
	w := g.Writer()
	w.Notice("tailing events every %s (ctrl-c to stop)…", c.Interval)
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()
	seen := map[string]bool{}
	for {
		select {
		case <-g.Context().Done():
			return nil
		case <-ticker.C:
		}
		q := url.Values{}
		q.Set("limit", strconv.Itoa(20))
		if c.EventType != "" {
			q.Set("event_type", c.EventType)
		}
		if c.ResourceType != "" {
			q.Set("resource_type", c.ResourceType)
		}
		out := struct {
			Data []map[string]any `json:"data"`
		}{}
		if err := execListSilent(g, "/events", q, &out); err != nil {
			w.Notice("poll error: %v (retrying)", err)
			continue
		}
		for i := len(out.Data) - 1; i >= 0; i-- {
			ev := out.Data[i]
			id, _ := ev["id"].(string)
			if id == "" || seen[id] {
				continue
			}
			seen[id] = true
			if err := renderLine(g, ev); err != nil {
				return fmt.Errorf("write event: %w", err)
			}
		}
	}
}
