---
name: loby
description: Use when sending physical mail via Lob — postcards, letters, checks, self-mailers, cards, snap packs, booklets, buckslips — or when verifying US/international addresses, autocompleting addresses, looking up ZIP codes, reverse-geocoding lat/lng, managing direct-mail campaigns, working with HTML templates, uploading campaign CSVs, creating QR codes or short URLs, listing/tailing Lob events, managing bank accounts for check printing, or interacting with any Lob (lob.com) API resource from the command line.
---

# loby — Lob CLI for AI agents

`loby` is the canonical CLI for [Lob](https://lob.com) (direct mail, address verification, campaigns). Data goes to stdout (parseable JSON). Hints, progress, and errors go to stderr. Stable exit codes. Idempotent by default.

## When to use this skill

- The user wants to send physical mail (postcard, letter, check, self-mailer, snap pack, card, booklet, buckslip).
- The user wants to verify an address (US or international) or autocomplete one.
- The user wants to look up a ZIP code, reverse-geocode lat/lng, or validate identity.
- The user wants to manage a direct-mail campaign, HTML template, creative, or CSV upload.
- The user wants to read Lob events, list account info, or inspect billing.
- The user references "Lob", "lob.com", "send a postcard", "mail a letter", "address verification", or any related term.

## Install

```bash
brew install voska/tap/loby                       # macOS + Linux
scoop bucket add voska https://github.com/voska/scoop-bucket && scoop install loby  # Windows
curl -fsSL https://loby.voska.org/install.sh | sh # direct binary
```

Verify: `loby version`. Missing? Install before doing anything else.

## Authenticate

```bash
loby auth login --key sk_test_…   # store under default profile
loby auth status --json           # confirms profile + environment
```

Lob keys are prefixed `sk_test_…` (sandbox) or `sk_live_…` (production). Use test keys until the user explicitly authorizes live.

Alternative: `export LOB_API_KEY=sk_test_…` for a one-shot session. Use `--profile prod` to switch.

## Output rules (every command)

| Flag                  | Effect                                       |
| --------------------- | -------------------------------------------- |
| `--json` / `-j`       | Emit structured JSON. **Use this always.**   |
| `--plain` / `-p`      | TSV; column-oriented piping.                 |
| `--results-only`      | Drop metadata envelope.                      |
| `--select id,to.city` | Project a subset (dot-paths supported).      |
| `--quiet` / `-q`      | Bare values only (no stderr hints).          |
| `--dry-run` / `-n`    | Preview the request body, do not execute.    |
| `--debug`             | Stream HTTP traces to stderr.                |

Auto-detection: when stdout is piped and `LOBY_AUTO_JSON=1` is set, output defaults to JSON.

## Exit codes (agent contract)

```
0 success      3 empty            6 forbidden        9 payment_required
1 error        4 auth_required    7 rate_limited     10 config_error
2 usage        5 not_found        8 retryable
```

Authoritative: `loby exit-codes --json`. Codes 7 and 8 are transient (retry with backoff).

## Introspection (use these to discover features)

```bash
loby schema --json                  # full command tree, every flag
loby schema postcards create --json # one command's signature
loby exit-codes --json              # exit codes
loby auth status --json             # active profile
loby version --json                 # build info
```

If you don't know a command, run `loby schema --json | jq` instead of guessing.

## Idempotency

Every mutating command auto-generates an `Idempotency-Key` and Lob caches the response for 24h. Retrying with the same flags returns the cached resource — exactly one postcard mails, not ten. Override with `--idempotency-key <key>`.

## Canonical recipes

**Send a postcard to a verified US address**

```bash
loby verify us "185 Berry St, San Francisco, CA 94107" --json --select deliverability
# proceed only if deliverability=deliverable
loby postcards create \
  --to '{"name":"Alice","address_line1":"185 Berry St","address_city":"San Francisco","address_state":"CA","address_zip":"94107"}' \
  --front @front.html \
  --back @back.html \
  --size 4x6 --json
```

**Verify a list of addresses (sync, up to 100)**

```bash
loby bulk us --addresses @addresses.json --json
```

**Send a check**

```bash
loby bank-accounts create --routing-number 122100024 --account-number 123456789 \
  --account-type company --signatory "Jane Doe" --json
loby checks create --to <adr_id> --bank-account <bank_id> --amount 250.00 \
  --memo "Invoice #1234" --json
```

**Stream events**

```bash
loby events tail --resource-type postcards --event-type postcard.created --json
```

**End-to-end campaign**

```bash
loby templates create --description "Promo Q3" --html @promo.html --json
# capture tmpl_id
loby campaigns create --name "Q3 Spring Promo" --schedule-type in_future \
  --send-date 2026-07-01 --json
# capture cmp_id
loby creatives create --campaign-id <cmp_id> --resource-type postcard \
  --from <adr_id> --details '{"front":"tmpl_…","back":"tmpl_…"}' --json
loby uploads create --campaign-id <cmp_id> --json
# capture upl_id
loby uploads file <upl_id> ./recipients.csv --json
loby uploads status <upl_id> --json   # poll until "verified"
loby campaigns send <cmp_id> --confirm --json
```

## Safety rules

- Always pass `--json`. The human formatter is for terminals, not agents.
- Always run `--dry-run` first when the user has not explicitly authorized the spend.
- Delete and cancel require `--confirm` (or `--force`). No exceptions.
- Never paste an API key into command flags in scripts — use `loby auth login` or `LOB_API_KEY`.
- Live keys (`sk_live_…`) charge real money. Default to test keys.

## Resources

- Full command catalog: [references/COMMANDS.md](references/COMMANDS.md)
- Verified recipes: [references/RECIPES.md](references/RECIPES.md)
- Lob resource glossary: [references/RESOURCES.md](references/RESOURCES.md)
- Upstream API: https://docs.lob.com
- Source: https://github.com/voska/loby
