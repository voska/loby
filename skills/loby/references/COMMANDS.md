# loby Command Catalog

This is a human-readable copy of `loby schema --json`. The binary itself is the source of truth — run `loby schema --json | jq` for the live tree.

## Global flags (every command)

```
-j, --json                Emit JSON to stdout         ($LOBY_JSON)
-p, --plain               Emit TSV to stdout          ($LOBY_PLAIN)
    --results-only        Drop the metadata envelope
    --select <fields>     Project output (comma-separated dot-paths)
    --no-color            Disable ANSI                ($NO_COLOR)
-q, --quiet               Suppress stderr hints
    --no-input            Fail rather than prompt
    --api-key <key>       Bypass keyring + env        ($LOB_API_KEY)
    --profile <name>      Named auth profile          ($LOB_PROFILE, default "default")
-n, --dry-run             Preview without executing
    --idempotency-key <k> Override the generated key
    --debug               Verbose stderr logging      ($LOBY_DEBUG)
-h, --help                Context-sensitive help
```

## Meta

| Command | Purpose |
| --- | --- |
| `loby version` | Build info (version, commit, date, runtime). |
| `loby schema [PATH]` | Full or scoped CLI tree as JSON. |
| `loby exit-codes` | Canonical exit-code table. |
| `loby auth login [--key]` | Store an API key under the active profile. |
| `loby auth logout` | Remove the active profile. |
| `loby auth status` | Active profile + key prefix + environment. |
| `loby auth list` | All profiles stored in the keyring. |
| `loby account` | Active Lob account info, balance, feature flags. |

## Addresses & verification

| Command | Purpose |
| --- | --- |
| `loby addresses create --line1 ... --name ...` | Save an address. |
| `loby addresses get <adr_id>` | Retrieve a saved address. |
| `loby addresses list` | Paginate saved addresses. |
| `loby addresses delete <adr_id> --confirm` | Delete a saved address. |
| `loby addresses verify "185 Berry St, …"` | Verify a US address. |
| `loby addresses autocomplete "185 Berr"` | Suggest completions. |
| `loby verify us "…"` | Same as `addresses verify`. |
| `loby verify intl "…" --country DE` | Verify an international address. |
| `loby zip 94107` | Look up city/state for a US ZIP. |
| `loby geo reverse --lat 37.78 --lng=-122.4` | Reverse-geocode lat/lng. |
| `loby identity verify --recipient "Larry Lobster" --primary-line "210 King St" --city … --state CA --zip 94107` | Validate identity at an address. |
| `loby bulk us --addresses @file.json` | Sync bulk US verification (≤100). |
| `loby bulk intl --addresses @file.json` | Sync bulk international verification. |

## Mail creation

Each resource has `create`, `get <id>`, and `list`. Cancellation support varies because Lob doesn't expose it uniformly:

| Resource | Verbs | Notes |
| --- | --- | --- |
| `loby postcards` | create, get, list | 4x6, 6x9, 6x11. No cancel — postcards enter USPS immediately on create. |
| `loby letters` | create, get, list, cancel | PDF, HTML, or template. Color, double-sided, perforated, certified mail. Cancel = `DELETE /letters/:id`. |
| `loby checks` | create, get, list, cancel | Requires a verified bank account. Amount, memo, message, logo. |
| `loby self-mailers` | create, get, list | 6x18 / 12x9 / 11x9 bifolds. No cancel via API. |
| `loby snap-packs` | create, get, list, cancel | 8.5x11 self-sealing snap packs. |

## Print assets (campaign artwork)

CRUD verbs: `create`, `get <id>`, `list`, `delete <id> --confirm`.

| Resource | Notes |
| --- | --- |
| `loby cards` | Card stock. |
| `loby booklets` | Multi-page booklets. |
| `loby buckslips` | 8.75x3.75 inserts. |

## Campaigns

| Command | Purpose |
| --- | --- |
| `loby campaigns create --name … --schedule-type immediate\|in_future` | Create campaign. |
| `loby campaigns get <cmp_id>` | Retrieve. |
| `loby campaigns list` | Paginate. |
| `loby campaigns update <cmp_id> …` | Update name/description/send_date. |
| `loby campaigns delete <cmp_id> --confirm` | Delete (only before send). |
| `loby campaigns send <cmp_id> --confirm` | Submit for processing. Irreversible. |
| `loby informed-delivery create --campaign-id …` | Informed Delivery campaign. |
| `loby creatives create --campaign-id … --resource-type postcard --from …` | Campaign artwork. |
| `loby uploads create --campaign-id …` | Upload metadata record. |
| `loby uploads file <upl_id> ./recipients.csv` | Attach the CSV. |
| `loby uploads get <upl_id>` | Retrieve upload (status is in the body). |
| `loby uploads exports create <upl_id> --type failures` | Generate row-error report. |
| `loby uploads exports list <upl_id>` | List export jobs. |
| `loby uploads exports get <upl_id> <ex_id>` | Retrieve a specific export. |
| `loby uploads report <upl_id>` | Line-item report (feature-flagged). |

## Templates

| Command | Purpose |
| --- | --- |
| `loby templates create --html @body.html` | Create with merge variables. |
| `loby templates get <tmpl_id>` | Retrieve. |
| `loby templates list` | Paginate. |
| `loby templates update <tmpl_id> --published-version <vrsn_id>` | Publish a version. |
| `loby templates delete <tmpl_id> --confirm` | Delete. |
| `loby templates versions create <tmpl_id> --html @body.html` | Create a version. |
| `loby templates versions get <tmpl_id> <vrsn_id>` | Retrieve. |
| `loby templates versions list <tmpl_id>` | Paginate versions. |

## Banking & billing

| Command | Purpose |
| --- | --- |
| `loby bank-accounts create --routing-number … --account-number … --signatory …` | Add account. |
| `loby bank-accounts verify <bank_id> --amounts 11,35` | Confirm test deposits. |
| `loby bank-accounts list` | Paginate. |
| `loby bank-accounts delete <bank_id> --confirm` | Delete. |
| `loby billing-groups create --name …` | Create group. |
| `loby billing-groups list` | Paginate. |

## URL shortener + QR analytics

Lob's URL shortener has two resources: `links` (short URLs) and `domains` (custom short-link domains). QR codes are minted by embedding Lob's snippet in mailer HTML; the API only surfaces scan analytics.

| Command | Purpose |
| --- | --- |
| `loby links create --redirect-link https://example.com/long`  | Create a tracked short URL. |
| `loby links get <link_id>` / `list` / `delete <link_id> --confirm` | Manage short URLs. |
| `loby domains create --domain links.example.com` | Register a custom short-link domain. |
| `loby domains get <id>` / `list` / `delete <id> --confirm` | Manage custom domains. |
| `loby qr-codes list [--scanned] [--limit N]` | QR code scan analytics. |

## Events + resource proofs

| Command | Purpose |
| --- | --- |
| `loby events list [--resource-type postcards] [--event-type postcard.created]` | Paginate events. |
| `loby events tail --interval 5s` | NDJSON stream of new events. |
| `loby events get <evt_id>` | Single event. |
| `loby resource-proofs get <id>` | PDF preview of a printed asset. |

## Argument conventions

- `--to` / `--from` accept either an address ID (`adr_…`) or inline JSON (`'{"name":"…"}'`) or `@file.json`.
- `--front`, `--back`, `--inside`, `--outside`, `--cover`, `--html`, `--file` accept a URL, a template ID (`tmpl_…`), inline HTML, or `@file.html`.
- `--metadata key=value` accepts repeated pairs (Lob limit: 20 keys, 500-char values).
- `--merge-variables` accepts inline JSON or `@file.json`.
