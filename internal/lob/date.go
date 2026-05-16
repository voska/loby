package lob

import (
	"strings"
	"time"
)

// Date is a flexible UnmarshalJSON wrapper for Lob fields that can come back
// either as a full RFC3339 timestamp or as a bare date string ("2026-05-16").
// Some Lob fields (e.g. expected_delivery_date, send_date, start_date) are
// documented as full timestamps but return date-only values for resources
// created with a test key — encoding/json's default *time.Time then rejects
// them. Date accepts both and renders as the original string.
type Date struct {
	t   time.Time
	raw string
}

// Time returns the underlying timestamp. Hour/minute/second are zero for
// date-only values.
func (d Date) Time() time.Time { return d.t }

// String returns the original JSON representation. Useful for round-tripping.
func (d Date) String() string { return d.raw }

// MarshalJSON emits the original string form, preserving date-only inputs.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.raw == "" {
		if d.t.IsZero() {
			return []byte(`null`), nil
		}
		return []byte(`"` + d.t.Format(time.RFC3339) + `"`), nil
	}
	return []byte(`"` + d.raw + `"`), nil
}

// UnmarshalJSON accepts null, an RFC3339 timestamp, or a bare YYYY-MM-DD date.
func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}
	d.raw = s
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		d.t = t
		return nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		d.t = t
		return nil
	}
	// Tolerate unknown formats — store the raw string, leave t zero.
	return nil
}
