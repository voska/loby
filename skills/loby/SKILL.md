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
curl -fsSL https://loby.voska.org/install.sh | sh
```

Verify: `loby version`.

## Authenticate

```bash
loby auth login --key sk_test_…
loby auth status --json
```

Lob keys are prefixed `sk_test_…` (sandbox) or `sk_live_…` (live). Use test keys until the user explicitly authorizes live. `LOB_API_KEY=…` works for one-shot use.

## Rules for agents (read these once)

1. **Always pass `--json`.** The human formatter is for terminals.
2. **Run `--dry-run` first** when the user has not explicitly authorized the spend.
3. **Delete and cancel require `--confirm` or `--force`.**
4. **Use `loby schema --json | jq` to discover** anything you don't know — it is the source of truth.
5. **Exit codes** drive control flow: `0` success, `3` empty, `4` auth, `5` not_found, `6` forbidden, `7` rate_limited (transient), `8` retryable (transient), `9` payment_required, `10` config_error. Authoritative table: `loby exit-codes --json`.
6. **Idempotency is automatic.** Every mutating command generates a deterministic `Idempotency-Key` from command + flags + body; Lob caches the response for 24h. Same flags → same response, exactly one mailed.

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

- `--to` / `--from` accept an address ID (`adr_…`), inline JSON, or `@file.json`.
- `--front`, `--back`, `--inside`, `--outside`, `--cover`, `--html`, `--file` accept HTML strings, URLs, template IDs (`tmpl_…`), or `@file`. Text files (.html, .csv, .md) ship as strings; binary files (.pdf, .png, .jpg) are base64-encoded as `data:` URIs.
- `--metadata key=value` accepts repeated pairs (Lob limit: 20 keys, 500-char values).
- `--merge-variables` accepts inline JSON or `@file.json`.
- `informed-delivery create` and other multipart endpoints take real files via `--ride-along-image @path.jpg`.

## Resources

Full command catalog: [references/COMMANDS.md](references/COMMANDS.md).
End-to-end recipes (postcard, check, bulk verify, full campaign, event tail): [references/RECIPES.md](references/RECIPES.md).
Lob resource glossary with ID prefixes: [references/RESOURCES.md](references/RESOURCES.md).
Upstream API docs: https://docs.lob.com. Source: https://github.com/voska/loby.

If something isn't in this skill, run `loby schema [subcommand] --json` and trust the binary over the docs.
