---
name: loby
description: Use when sending physical mail via Lob — postcards, letters, checks, self-mailers, cards, snap packs, booklets, buckslips — or when verifying US/international addresses, autocompleting addresses, looking up ZIP codes, reverse-geocoding lat/lng, managing direct-mail campaigns, working with HTML templates, uploading campaign CSVs, creating QR codes or short URLs, listing/tailing Lob events, managing bank accounts for check printing, or interacting with any Lob (lob.com) API resource from the command line.
---

# loby — Lob CLI for AI agents

`loby` is the canonical CLI for [Lob](https://lob.com) (direct mail, address verification, campaigns). Data → stdout (parseable JSON). Hints, progress, errors → stderr. Stable exit codes. Idempotent by default.

## Install

```bash
brew install voska/tap/loby                       # macOS + Linux
scoop bucket add voska https://github.com/voska/scoop-bucket && scoop install loby
curl -fsSL https://lobycli.com/install.sh | sh
```

Verify: `loby version`.

## Authenticate

```bash
loby auth login --key test_…
loby auth status --json
```

Lob keys are prefixed `test_…` (sandbox) or `live_…` (production); publishable keys (limited to address verification, autocompletion, and ZIP/geo lookups) use `test_pub_…` or `live_pub_…`. Use test keys until the user explicitly authorizes live. `LOB_API_KEY=…` works for one-shot use. Live keys 401 with `invalid_api_key` until the account has verified its email AND added a payment method — adding Lob Credits is not sufficient.

## Rules for agents (read these once)

1. **Always pass `--json`.** The human formatter is for terminals.
2. **Run `--dry-run` first** when the user has not explicitly authorized the spend.
3. **Delete and cancel require `--confirm` or `--force`.**
4. **Use `loby schema --json | jq` to discover** anything you don't know — it is the source of truth.
5. **Exit codes** drive control flow: `0` success, `3` empty, `4` auth, `5` not_found, `6` forbidden, `7` rate_limited (transient), `8` retryable (transient), `9` payment_required, `10` config_error. Authoritative table: `loby exit-codes --json`.
6. **Idempotency is automatic on mailer creates.** Every mutating command generates a deterministic `Idempotency-Key` from command + flags + body. Lob caches the response for 24h on mail-creation endpoints (postcards, letters, checks, self-mailers, snap-packs, etc.), so retrying with the same flags returns the same resource ID. Address book, links, templates, and other utility endpoints do *not* cache — those are inherently safe to retry because they're either idempotent already (PUT/DELETE) or cheap to deduplicate client-side. Override the auto-key with `--idempotency-key <key>` when you want explicit control.

## Output

| Flag | Effect |
| --- | --- |
| `--json` / `-j` | Structured JSON. Default for agents. |
| `--plain` / `-p` | TSV; column-oriented piping. |
| `--results-only` | Strip `{data:[...]}` envelope on list responses. |
| `--select id,to.city` | Project fields with dot-paths. |
| `--quiet` / `-q` | Emit bare IDs only (one per line). |
| `--dry-run` / `-n` | Preview the request; do not execute mutations. |

`LOBY_AUTO_JSON=1` makes piped invocations default to JSON. `--no-input` is implied when stdin is not a TTY.

## Argument conventions

- `--to` / `--from` accept an address ID (`adr_…`), inline JSON `{"name":"…","address_line1":"…",…}`, or `@file.json`.
- `--front`, `--back`, `--inside`, `--outside`, `--cover`, `--html`, `--file` accept HTML strings, URLs, template IDs (`tmpl_…`), or `@file`. Text files (.html, .csv, .md) ship as strings; binary files (.pdf, .png, .jpg) are base64-encoded as `data:` URIs.
- `--metadata key=value` accepts repeated pairs (Lob limit: 20 keys, 500-char values).
- `--merge-variables` accepts inline JSON or `@file.json`.
- `informed-delivery create` and other multipart endpoints take real files via `--ride-along-image @path.jpg`.
- `geo reverse` takes `--lat` and `--lng` (use `--lng=-122.4` for negative longitudes so the parser doesn't read `-` as a short flag).
- `identity verify` takes `--recipient` (full name) or `--company`, plus a US address via `--primary-line` etc.
- `creatives create` is the exception to the HTML rule: Lob rejects inline HTML on `/v1/creatives` and only accepts PDF URLs or `tmpl_…` template IDs on `--front/--back/--cover/--file`. Postcards/letters/etc. accept inline HTML; campaign creatives don't.

## Surface limits

- **Campaigns** are write-once at create time — no update verb exists in the API.
- **Creatives** are POST-only (`/v1/creatives`); they have no get, list, update, or delete.
- **Upload exports** support create + get only — there is no listing endpoint. The `create` response returns the export ID in `.exportId` (camelCase), not `.id`.

## Cancel semantics

- `letters`, `checks`, `snap-packs` support `cancel <id> --confirm` (issues `DELETE /<type>/:id` to Lob).
- `postcards` and `self-mailers` enter USPS on create and cannot be cancelled via the API.

## Resources

Full command catalog: [references/COMMANDS.md](references/COMMANDS.md).
End-to-end recipes (postcard, check, bulk verify, full campaign, event tail): [references/RECIPES.md](references/RECIPES.md).
Lob resource glossary with ID prefixes: [references/RESOURCES.md](references/RESOURCES.md).
Upstream API docs: https://docs.lob.com. Source: https://github.com/voska/loby.

If something isn't in this skill, run `loby schema [subcommand] --json` and trust the binary over the docs.
