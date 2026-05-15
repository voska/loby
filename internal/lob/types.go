// Package lob holds typed request/response shapes for the Lob v1 API. These
// types are the JSON contract loby exposes to humans and agents — changes are
// schema-breaking.
package lob

import "time"

// Metadata is Lob's per-resource key/value bag (string values only, ≤500 char,
// ≤20 keys per resource).
type Metadata map[string]string

// List is the standard paginated response shape used across Lob list endpoints.
type List[T any] struct {
	Object      string `json:"object"`
	Data        []T    `json:"data"`
	NextURL     string `json:"next_url,omitempty"`
	PreviousURL string `json:"previous_url,omitempty"`
	Count       int    `json:"count"`
	TotalCount  *int   `json:"total_count,omitempty"`
}

// Deleted is the response shape Lob returns from DELETE endpoints.
type Deleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// MailType is the Lob delivery class. Used on postcards, letters, checks, etc.
type MailType string

// Recognized MailType values per Lob's spec.
const (
	MailTypeUSPSFirstClass MailType = "usps_first_class"
	MailTypeUSPSStandard   MailType = "usps_standard"
)

// Tracking holds the carrier tracking info on mailed items.
type Tracking struct {
	ID             string     `json:"id,omitempty"`
	TrackingNumber string     `json:"tracking_number,omitempty"`
	Carrier        string     `json:"carrier,omitempty"`
	Object         string     `json:"object,omitempty"`
	DateCreated    *time.Time `json:"date_created,omitempty"`
	DateModified   *time.Time `json:"date_modified,omitempty"`
}

// Thumbnail is the rendered preview shape Lob attaches to most mailers.
type Thumbnail struct {
	Small  string `json:"small,omitempty"`
	Medium string `json:"medium,omitempty"`
	Large  string `json:"large,omitempty"`
}
