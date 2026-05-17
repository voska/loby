# loby

> The canonical CLI for [Lob](https://lob.com) — direct mail, address verification, and campaigns. For humans and AI agents.

[![CI](https://github.com/voska/loby/actions/workflows/ci.yml/badge.svg)](https://github.com/voska/loby/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/voska/loby?logo=github)](https://github.com/voska/loby/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/voska/loby.svg)](https://pkg.go.dev/github.com/voska/loby)
[![MIT](https://img.shields.io/github/license/voska/loby)](LICENSE)

```text
$ loby verify us "185 Berry St, San Francisco, CA 94107" --json --select deliverability
{"deliverability": "deliverable"}

$ loby postcards create --to adr_… --front @front.html --back @back.html --json
{"id": "psc_abc123", "status": "rendered", "carrier": "USPS", "expected_delivery_date": "…"}
```

## Why

Lob has SDKs in seven languages and zero canonical CLIs. AI agents need one tool, one binary, one install command, and stable structured output to drive Lob's API end-to-end. `loby` is that tool.

- **Single static binary** — Go, zero runtime dependencies.
- **All 29 Lob resources** — postcards, letters, checks, self-mailers, snap packs, cards, booklets, buckslips, addresses, verifications (US, intl, bulk, CSV), autocompletion, ZIP/reverse-geo lookups, identity validation, campaigns, informed delivery, creatives, uploads, templates, bank accounts, billing groups, QR codes, short URLs, events, resource proofs, accounts.
- **Agent-first** — structured JSON on stdout, hints on stderr, stable exit codes, full schema introspection, idempotent by default, `--dry-run` on every mutation.
- **Boring stack** — Kong CLI, OS keychain (99designs/keyring), hand-rolled `net/http` for full control. No third-party HTTP client, no junk dependencies.

## Install

```bash
# macOS + Linux (Homebrew)
brew install voska/tap/loby

# Windows (Scoop)
scoop bucket add voska https://github.com/voska/scoop-bucket
scoop install loby

# Direct binary (any platform)
curl -fsSL https://lobycli.com/install.sh | sh

# From source
go install github.com/voska/loby/cmd/loby@latest
```

Verify: `loby version`.

## Authenticate

```bash
loby auth login                          # interactive prompt
loby auth login --key test_…          # one-shot
loby auth status --json                  # confirm
```

Keys live in the OS keychain (Keychain on macOS, secret-service on Linux, Credential Manager on Windows). Override per-invocation with `--api-key` or `LOB_API_KEY`. Use named profiles for test vs live: `loby auth login --profile prod`.

## Use

```bash
# Send a postcard end-to-end.
loby verify us "185 Berry St, San Francisco, CA 94107" --json --select deliverability
loby postcards create \
  --to '{"name":"Alice","address_line1":"185 Berry St","address_city":"San Francisco","address_state":"CA","address_zip":"94107"}' \
  --front @front.html --back @back.html \
  --size 4x6 --json

# Bulk-verify addresses.
loby bulk us --addresses @addresses.json --json

# Manage a campaign.
loby campaigns create --name "Q3 Promo" --schedule-type in_future --send-date 2026-07-01 --json
loby uploads create --campaign-id cmp_… --json
loby uploads file <upl_id> ./recipients.csv --json
loby campaigns send cmp_… --confirm --json

# Stream events.
loby events tail --resource-type postcards --json | tee events.ndjson
```

Every command supports `--dry-run`, `--json`, `--plain`, `--select`, `--profile`, `--idempotency-key`. Discover everything:

```bash
loby schema --json                  # full CLI tree
loby schema postcards create --json # one command
loby exit-codes --json              # exit codes
loby --help                         # human-readable
```

## For AI agents

Drop `skills/loby/SKILL.md` into your skills directory:

```bash
mkdir -p ~/.claude/skills/loby
curl -fsSL https://lobycli.com/skill/SKILL.md > ~/.claude/skills/loby/SKILL.md
```

Or install the full bundle (SKILL.md + references):

```bash
curl -fsSL https://lobycli.com/skill/install.sh | sh
```

The skill teaches agents to install the CLI, authenticate, prefer `--json`, use `--dry-run`, and follow the canonical mail-flow recipes.

## Output

| Mode | Trigger | Use |
| --- | --- | --- |
| Human | default TTY | Colored tables and key/value views. |
| JSON | `--json` / `-j` or `LOBY_JSON=1` | Structured output for parsing. |
| Plain | `--plain` / `-p` | Tab-separated values. |
| NDJSON | `events tail`, list streaming | One JSON object per line. |

Field projection: `--select id,to.city,status`. Pipe-detection: `LOBY_AUTO_JSON=1 loby … | jq` automatically switches to JSON when stdout isn't a TTY.

## Exit codes

```
0 success      3 empty            6 forbidden        9 payment_required
1 error        4 auth_required    7 rate_limited     10 config_error
2 usage        5 not_found        8 retryable
```

Authoritative: `loby exit-codes --json`. Codes 7 and 8 are transient — automatic backoff retries are built in.

## Idempotency

Every mutating command auto-generates a deterministic `Idempotency-Key` from `command + flags + body`. Lob caches the response for 24h on mail-creation endpoints (postcards, letters, checks, self-mailers, snap-packs, cards/booklets/buckslips, campaigns/creatives/uploads). Retrying a mailer create with the same flags returns the same resource ID — one postcard mails, exactly. Utility endpoints (address book, links, templates, lookups) don't replay; the key is sent but Lob produces a fresh ID each call. Override the key with `--idempotency-key <key>`.

## Develop

```bash
git clone https://github.com/voska/loby
cd loby
make ci          # fmt-check + vet + lint + test + build
make build       # bin/loby
make test        # unit
make test-integration  # requires LOB_API_KEY=test_…
```

See [AGENTS.md](AGENTS.md) for the build contract and [docs/superpowers/specs/2026-05-15-loby-design.md](docs/superpowers/specs/2026-05-15-loby-design.md) for the spec.

## License

MIT © 2026 Matt Voska.

Not affiliated with Lob, Inc. "Lob" is a trademark of Lob, Inc.
