package cli

// Root is the top-level Kong CLI definition. Each *Cmd field becomes a
// subcommand; new resource groups are added by appending a field here and
// dropping a Kong struct alongside.
type Root struct {
	Globals

	// Introspection / meta
	Version   VersionCmd   `cmd:"" help:"Print build information."`
	Schema    SchemaCmd    `cmd:"" help:"Print the CLI command tree as JSON (agent introspection)."`
	ExitCodes ExitCodesCmd `cmd:"" help:"Print the canonical exit-code table." aliases:"exitcodes,exit"`
	Auth      AuthCmd      `cmd:"" help:"Manage Lob API credentials."`

	// Address & verification
	Addresses AddressesCmd `cmd:"" help:"Manage saved addresses, verify, autocomplete."`
	Verify    VerifyCmd    `cmd:"" help:"Verify US or international addresses."`
	Zip       ZipCmd       `cmd:"" help:"Look up city/state for a US ZIP code."`
}
