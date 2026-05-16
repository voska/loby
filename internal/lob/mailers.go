package lob

import "time"

// MailerCore is the common shape across all Lob mail objects (postcards,
// letters, checks, self_mailers, cards, booklets, buckslips, snap_packs).
// Each typed mailer embeds this and adds resource-specific fields.
//
// Date-shaped fields use lob.Date — Lob documents them as full timestamps but
// returns bare YYYY-MM-DD for many resources, especially under test keys.
type MailerCore struct {
	ID               string         `json:"id"`
	Description      string         `json:"description,omitempty"`
	Metadata         Metadata       `json:"metadata,omitempty"`
	MergeVariables   map[string]any `json:"merge_variables,omitempty"`
	MailType         MailType       `json:"mail_type,omitempty"`
	URL              string         `json:"url,omitempty"`
	Carrier          string         `json:"carrier,omitempty"`
	Tracking         *Tracking      `json:"tracking_number,omitempty"`
	TrackingEvents   []any          `json:"tracking_events,omitempty"`
	Thumbnails       []Thumbnail    `json:"thumbnails,omitempty"`
	ExpectedDelivery Date           `json:"expected_delivery_date,omitempty"`
	DateCreated      time.Time      `json:"date_created"`
	DateModified     time.Time      `json:"date_modified"`
	SendDate         Date           `json:"send_date,omitempty"`
	Status           string         `json:"status,omitempty"`
	Object           string         `json:"object,omitempty"`
	UseType          string         `json:"use_type,omitempty"`
}

// Postcard is the GET /v1/postcards/:id response.
type Postcard struct {
	MailerCore
	To    any    `json:"to,omitempty"`
	From  any    `json:"from,omitempty"`
	Front string `json:"front,omitempty"`
	Back  string `json:"back,omitempty"`
	Size  string `json:"size,omitempty"`
}

// Letter is the GET /v1/letters/:id response.
type Letter struct {
	MailerCore
	To               any    `json:"to,omitempty"`
	From             any    `json:"from,omitempty"`
	File             string `json:"file,omitempty"`
	Color            bool   `json:"color"`
	DoubleSided      bool   `json:"double_sided"`
	AddressPlacement string `json:"address_placement,omitempty"`
	ReturnEnvelope   any    `json:"return_envelope,omitempty"`
	PerforatedPage   *int   `json:"perforated_page,omitempty"`
	CustomEnvelope   any    `json:"custom_envelope,omitempty"`
	ExtraService     string `json:"extra_service,omitempty"`
}

// Check is the GET /v1/checks/:id response. BankAccount comes back as an
// inflated object on retrieve, not just an ID — typed as `any` to accept both.
type Check struct {
	MailerCore
	To            any     `json:"to,omitempty"`
	From          any     `json:"from,omitempty"`
	BankAccount   any     `json:"bank_account,omitempty"`
	CheckNumber   int     `json:"check_number,omitempty"`
	Memo          string  `json:"memo,omitempty"`
	Message       string  `json:"message,omitempty"`
	Logo          any     `json:"logo,omitempty"`
	Amount        float64 `json:"amount"`
	AttachmentURL any     `json:"attachment,omitempty"`
	Pages         int     `json:"pages,omitempty"`
}

// SelfMailer is the GET /v1/self_mailers/:id response.
type SelfMailer struct {
	MailerCore
	To      any    `json:"to,omitempty"`
	From    any    `json:"from,omitempty"`
	Outside string `json:"outside,omitempty"`
	Inside  string `json:"inside,omitempty"`
	Size    string `json:"size,omitempty"`
}

// Card is the GET /v1/cards/:id response (printed card stock).
type Card struct {
	ID              string    `json:"id"`
	Description     string    `json:"description,omitempty"`
	Front           string    `json:"front,omitempty"`
	Back            string    `json:"back,omitempty"`
	Size            string    `json:"size,omitempty"`
	AutoReorder     bool      `json:"auto_reorder,omitempty"`
	ReorderQuantity int       `json:"reorder_quantity,omitempty"`
	RawCost         string    `json:"raw_cost,omitempty"`
	Stock           string    `json:"stock,omitempty"`
	Status          string    `json:"status,omitempty"`
	Metadata        Metadata  `json:"metadata,omitempty"`
	DateCreated     time.Time `json:"date_created"`
	DateModified    time.Time `json:"date_modified"`
	Object          string    `json:"object,omitempty"`
}

// Booklet is the GET /v1/booklets/:id response.
type Booklet struct {
	ID           string    `json:"id"`
	Description  string    `json:"description,omitempty"`
	Inside       string    `json:"inside,omitempty"`
	Cover        string    `json:"cover,omitempty"`
	Size         string    `json:"size,omitempty"`
	Status       string    `json:"status,omitempty"`
	Metadata     Metadata  `json:"metadata,omitempty"`
	DateCreated  time.Time `json:"date_created"`
	DateModified time.Time `json:"date_modified"`
	Object       string    `json:"object,omitempty"`
}

// Buckslip and SnapPack mirror Card/Booklet's flat shape.
type Buckslip struct {
	ID              string    `json:"id"`
	Description     string    `json:"description,omitempty"`
	Front           string    `json:"front,omitempty"`
	Back            string    `json:"back,omitempty"`
	Size            string    `json:"size,omitempty"`
	Status          string    `json:"status,omitempty"`
	AutoReorder     bool      `json:"auto_reorder,omitempty"`
	ReorderQuantity int       `json:"reorder_quantity,omitempty"`
	Stock           string    `json:"stock,omitempty"`
	Metadata        Metadata  `json:"metadata,omitempty"`
	DateCreated     time.Time `json:"date_created"`
	DateModified    time.Time `json:"date_modified"`
	Object          string    `json:"object,omitempty"`
}

// SnapPack is the GET /v1/snap_packs/:id response.
type SnapPack struct {
	MailerCore
	To      any    `json:"to,omitempty"`
	From    any    `json:"from,omitempty"`
	Outside string `json:"outside,omitempty"`
	Inside  string `json:"inside,omitempty"`
	Size    string `json:"size,omitempty"`
	Color   bool   `json:"color,omitempty"`
}
