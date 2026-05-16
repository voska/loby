package cli

// Root is the top-level Kong CLI definition. Each *Cmd field becomes a
// subcommand; new resource groups are added by appending a field here and
// dropping a Kong struct alongside.
type Root struct {
	Globals

	// Introspection / meta
	Version    VersionCmd    `cmd:"" help:"Print build information."`
	Schema     SchemaCmd     `cmd:"" help:"Print the CLI command tree as JSON (agent introspection)."`
	ExitCodes  ExitCodesCmd  `cmd:"" help:"Print the canonical exit-code table." aliases:"exitcodes,exit"`
	Completion CompletionCmd `cmd:"" help:"Generate shell completion (bash, zsh, fish, powershell)."`
	Auth       AuthCmd       `cmd:"" help:"Manage Lob API credentials."`
	Account    AccountCmd    `cmd:"" help:"Show the active Lob account profile and balance."`

	// Address & verification
	Addresses AddressesCmd `cmd:"" help:"Manage saved addresses, verify, autocomplete."`
	Verify    VerifyCmd    `cmd:"" help:"Verify US or international addresses."`
	Zip       ZipCmd       `cmd:"" help:"Look up city/state for a US ZIP code."`
	Geo       GeoCmd       `cmd:"" help:"Reverse-geocode lat/lng to ZIP codes."`
	Identity  IdentityCmd  `cmd:"" help:"Verify an individual's identity at an address."`
	Bulk      BulkCmd      `cmd:"" help:"Bulk verification operations (sync arrays and async CSV jobs)."`

	// Mail creation
	Postcards   PostcardsCmd   `cmd:"" help:"Create, retrieve, list, and cancel postcards."`
	Letters     LettersCmd     `cmd:"" help:"Create, retrieve, list, and cancel letters."`
	Checks      ChecksCmd      `cmd:"" help:"Create, retrieve, list, and cancel checks."`
	SelfMailers SelfMailersCmd `cmd:"" help:"Create, retrieve, list, and cancel self-mailers." name:"self-mailers"`
	SnapPacks   SnapPacksCmd   `cmd:"" help:"Create, retrieve, list, and cancel snap packs." name:"snap-packs"`

	// Print assets (campaign artwork)
	Cards     CardsCmd     `cmd:"" help:"Manage card stock artwork."`
	Booklets  BookletsCmd  `cmd:"" help:"Manage booklet artwork."`
	Buckslips BuckslipsCmd `cmd:"" help:"Manage buckslip artwork."`

	// Campaigns
	Campaigns        CampaignsCmd        `cmd:"" help:"Manage direct-mail campaigns."`
	InformedDelivery InformedDeliveryCmd `cmd:"" help:"Manage USPS Informed Delivery campaigns." name:"informed-delivery"`
	Creatives        CreativesCmd        `cmd:"" help:"Manage campaign creatives."`
	Uploads          UploadsCmd          `cmd:"" help:"Manage campaign CSV uploads."`

	// Templates
	Templates TemplatesCmd `cmd:"" help:"Manage reusable HTML templates."`

	// Bank accounts & billing
	BankAccounts  BankAccountsCmd  `cmd:"" help:"Manage bank accounts for check mailing." name:"bank-accounts"`
	BillingGroups BillingGroupsCmd `cmd:"" help:"Manage billing groups." name:"billing-groups"`

	// QR analytics, URL shortener, events, proofs
	QRCodes        QRCodesCmd        `cmd:"" help:"List QR code scan analytics." name:"qr-codes"`
	Links          LinksCmd          `cmd:"" help:"Manage short links (Lob's URL shortener)."`
	Domains        DomainsCmd        `cmd:"" help:"Manage custom short-link domains."`
	Events         EventsCmd         `cmd:"" help:"List and tail Lob events."`
	ResourceProofs ResourceProofsCmd `cmd:"" help:"Retrieve resource proof previews." name:"resource-proofs"`
}
