package lob

import "time"

// Address is a Lob saved address record (object="address").
type Address struct {
	ID             string    `json:"id"`
	Description    string    `json:"description,omitempty"`
	Name           string    `json:"name,omitempty"`
	Company        string    `json:"company,omitempty"`
	Email          string    `json:"email,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	AddressLine1   string    `json:"address_line1"`
	AddressLine2   string    `json:"address_line2,omitempty"`
	AddressCity    string    `json:"address_city,omitempty"`
	AddressState   string    `json:"address_state,omitempty"`
	AddressZip     string    `json:"address_zip,omitempty"`
	AddressCountry string    `json:"address_country,omitempty"`
	Metadata       Metadata  `json:"metadata,omitempty"`
	DateCreated    time.Time `json:"date_created"`
	DateModified   time.Time `json:"date_modified"`
	Object         string    `json:"object,omitempty"`
}

// AddressCreate is the POST /v1/addresses request body. Either an inline
// address record or a reference to an existing address ID is accepted; this
// type is the inline variant.
type AddressCreate struct {
	Description    string   `json:"description,omitempty"`
	Name           string   `json:"name,omitempty"`
	Company        string   `json:"company,omitempty"`
	Email          string   `json:"email,omitempty"`
	Phone          string   `json:"phone,omitempty"`
	AddressLine1   string   `json:"address_line1"`
	AddressLine2   string   `json:"address_line2,omitempty"`
	AddressCity    string   `json:"address_city,omitempty"`
	AddressState   string   `json:"address_state,omitempty"`
	AddressZip     string   `json:"address_zip,omitempty"`
	AddressCountry string   `json:"address_country,omitempty"`
	Metadata       Metadata `json:"metadata,omitempty"`
}
