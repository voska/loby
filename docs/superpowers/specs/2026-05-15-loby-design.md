# loby — Lob CLI Design Spec

**Status:** Draft v1
**Date:** 2026-05-15
**Owner:** matt@voska.org
**Repo (planned):** github.com/voska/loby

## Mission

Ship the canonical 2026 CLI for Lob (direct mail) that AI agents can discover, install, and use to drive every Lob service end-to-end. Humans get a delightful tool; agents get a parseable, introspectable, idempotent API.

**Goal property:** `brew install voska/tap/loby` followed by a single SKILL.md drop is enough for an agent to send a postcard to a verified US address without reading Lob's docs.

## Scope (V1.0.0)

Full coverage of all 29 Lob resources, exposed via consistent CRUD + resource-specific verbs:

accounts, addresses, bank_accounts, billing_groups, booklets, buckslips, bulk_intl_verifications, bulk_us_verifications, campaigns, cards, checks, creatives, events, identity_validation, informed_delivery_campaigns, intl_verifications, letters, postcards, qr_codes, resource_proofs, reverse_geocode_lookups, self_mailers, snap_packs, templates, uploads, url_shortener, us_autocompletions, us_csv_verifications, us_verifications, zip_lookups.

Out of scope: webhooks ingest server, GUI, MCP server (CLI is the canonical interface; MCP is a thin shim someone else can build).

## Architectural decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | Language: Go | Single static binary, fast startup, goreleaser → brew/scoop, matches steipete/voska pattern |
| 2 | Binary name: `loby` | Distinct from English verb "lob" (search-recall); zero collision risk if Lob ships an official CLI; brandable |
| 3 | API client: hand-rolled `net/http` against `lob-api-public.yml` | `lob-go` covers only 23/29 resources, imports deprecated `ioutil`, drags in oauth2/protobuf/appengine for an HTTP-Basic-Auth API. Hand-rolled gives full coverage, zero junk deps, total control of error mapping. |
| 4 | CLI framework: `github.com/alecthomas/kong` | Struct-tag based; clean Go surface; easy `schema --json` generation from AST |
| 5 | Auth store: `github.com/99designs/keyring` | OS keychain primary; encrypted file fallback for headless |
| 6 | Config: TOML at `$XDG_CONFIG_HOME/loby/config.toml`; profiles for test/live | Standard XDG; one binary, many environments |
| 7 | Idempotency: auto-generated `Idempotency-Key` on every create; Lob caches responses server-side for 24h | Agents retry constantly; Lob already handles replay — local cache is YAGNI |
| 8 | Release: goreleaser → Homebrew tap + Scoop bucket + GitHub Releases (signed) | SV-standard distribution |
| 9 | Discovery: GH Pages site + `llms.txt` + `install.sh` one-liner + packaged SKILL.md | Designed for agent crawl + install |

## Repo layout

```
loby/
├── cmd/loby/main.go              # ~30 lines: version embed + Kong dispatch
├── internal/
│   ├── cli/                      # Kong command structs (one file per resource group)
│   │   ├── root.go               # top-level CLI struct, global flags
│   │   ├── auth.go               # auth login/logout/status/switch
│   │   ├── config.go             # config get/set/list
│   │   ├── schema.go             # schema --json (introspection)
│   │   ├── addresses.go          # addresses + us_autocompletions + zip_lookups
│   │   ├── postcards.go
│   │   ├── letters.go
│   │   ├── checks.go
│   │   ├── self_mailers.go
│   │   ├── cards.go
│   │   ├── booklets.go
│   │   ├── buckslips.go
│   │   ├── snap_packs.go
│   │   ├── templates.go          # templates + template_versions
│   │   ├── campaigns.go          # campaigns + informed_delivery_campaigns
│   │   ├── creatives.go
│   │   ├── verify.go             # us_verifications + intl_verifications + bulk_*
│   │   ├── identity.go           # identity_validation
│   │   ├── events.go             # events list/tail (NDJSON)
│   │   ├── qr.go                 # qr_codes + url_shortener
│   │   ├── geo.go                # reverse_geocode_lookups
│   │   ├── uploads.go
│   │   ├── resource_proofs.go
│   │   ├── bank_accounts.go      # bank_accounts + verify
│   │   ├── billing_groups.go
│   │   └── account.go            # accounts
│   ├── client/                   # Hand-rolled HTTP client (net/http)
│   │   ├── client.go             # New(profile) — base URL, auth header, idempotency injection
│   │   ├── do.go                 # Generic Do[T](ctx, method, path, body) -> (T, *Response, error)
│   │   ├── retry.go              # exponential backoff on 429 + 5xx (respects Retry-After)
│   │   ├── idempotency.go        # generate + persist + replay-key on retry
│   │   └── errors.go             # parse Lob error envelope → typed *APIError
│   ├── lob/                      # Resource types + per-resource methods (one file per resource)
│   │   ├── addresses.go
│   │   ├── postcards.go
│   │   └── ...                   # 29 files, one per Lob resource
│   ├── lobspec/                  # codegen artifacts from lob-api-public.yml (types only)
│   ├── output/                   # Formatters
│   │   ├── writer.go             # mode detection (TTY, --json, --plain, --results-only)
│   │   ├── human.go              # colored tables (termenv)
│   │   ├── json.go
│   │   ├── plain.go              # TSV
│   │   ├── ndjson.go             # streaming
│   │   └── select.go             # --select field projection (dot-path)
│   ├── auth/                     # Keyring + env + profile
│   │   ├── store.go              # 99designs/keyring abstraction
│   │   ├── env.go                # LOB_API_KEY, LOB_PROFILE
│   │   └── profile.go            # test vs live + named profiles
│   ├── config/                   # TOML config + XDG
│   ├── schema/                   # CLI tree → JSON (from Kong AST)
│   ├── errfmt/                   # error mapping → exit codes + recovery hints
│   └── version/                  # embedded version/commit/date
├── skills/loby/                  # Canonical agent skill (drop-in for ~/.claude/skills/)
│   ├── SKILL.md
│   ├── references/
│   │   ├── COMMANDS.md           # full command catalog
│   │   ├── RECIPES.md            # verified mail-flow recipes
│   │   └── RESOURCES.md          # Lob resource glossary
│   └── install.sh                # curl | bash installer for the skill bundle
├── site/                         # GH Pages
│   ├── index.html
│   ├── llms.txt                  # agent crawl manifest
│   └── install.sh                # `curl -fsSL loby.voska.org/install.sh | sh`
├── docs/
│   ├── auth.md
│   ├── distribution.md
│   └── recipes/                  # human-readable recipe docs (mirror skill/RECIPES.md)
├── scripts/
│   ├── live-test.sh              # smoke test against Lob test env
│   └── update-schema-snapshot.sh
├── .github/workflows/
│   ├── ci.yml                    # fmt-check, lint, test, build
│   ├── release.yml               # goreleaser on tag
│   └── pages.yml                 # deploy site/ on main
├── AGENTS.md                     # repo build/test/commit contract
├── CLAUDE.md                     # symlink → AGENTS.md
├── SPEC.md                       # symlink → docs/superpowers/specs/<this file>
├── README.md                     # human-facing
├── CHANGELOG.md
├── LICENSE                       # MIT
├── Makefile
├── .golangci.yml                 # strict v2
├── .goreleaser.yaml              # brew + scoop + GitHub Releases
├── go.mod / go.sum
```

## Command surface

### Global flags (every command)

| Flag | Env | Description |
|------|-----|-------------|
| `--json / -j` | `LOBY_JSON=1` | Structured JSON to stdout |
| `--plain / -p` | `LOBY_PLAIN=1` | TSV to stdout |
| `--results-only` | — | Strip metadata envelope |
| `--select <fields>` | — | Dot-path field projection |
| `--no-color` | `NO_COLOR` | Disable ANSI |
| `--no-input` | — | Fail rather than prompt |
| `--profile <name>` | `LOB_PROFILE` | Named auth profile (default: `default`) |
| `--api-key <key>` | `LOB_API_KEY` | Bypass keyring (CI use) |
| `--dry-run / -n` | — | Preview as JSON, don't execute (mutations only) |
| `--idempotency-key <k>` | — | Override auto-generated key |
| `--quiet / -q` | — | Print bare ID/status only |
| `--debug` | `LOBY_DEBUG=1` | Verbose stderr logging |

Auto-detection: when stdout is not a TTY, default mode becomes JSON if `LOBY_AUTO_JSON=1`.

### Resource verbs (uniform across resources where applicable)

```
loby <res> create [flags]              # POST
loby <res> get <id>                    # GET single
loby <res> list [--limit] [--before] [--after] [--include]
loby <res> delete <id> --confirm       # DELETE (destructive: requires --confirm or --force)
loby <res> cancel <id>                 # mail resources: pre-mailing cancel
```

### Resource-specific verbs (selected)

```
loby addresses verify <line1>... [--country US]
loby addresses autocomplete <prefix> [--state] [--city]
loby zip <zipcode>
loby geo reverse <lat> <lng>

loby postcards create --to <addr_id|json> --from <addr_id|json> \
  --front <html|url|@file.html> --back <html|url|@file.html> \
  [--mailing-date YYYY-MM-DD] [--size 4x6|6x9|6x11] [--description]
loby letters create --to ... --from ... --file <pdf|html|@file> [--color] [--double-sided]
loby checks create --to ... --bank-account <id> --amount 12.34 [--memo] [--logo @file.png]
loby self-mailers create --to ... --outside ... --inside ...
loby cards create --front --back --size
loby booklets create --inside <pdf> --cover ...
loby buckslips create ...
loby snap-packs create ...

loby templates create --description <s> --html <html|@file>
loby templates render <id> --merge-vars '{"name":"..."}' [--out preview.pdf]
loby templates versions <id>

loby campaigns create --name <s> --schedule-type immediate|in_future
loby campaigns send <id>
loby creatives create --campaign-id <id> --resource-type postcard --front --back

loby verify us '<addr>'                                          # shortcut
loby verify intl '<addr>' --country DE
loby verify bulk-us --file addresses.csv --concurrency 8         # async submission
loby verify bulk-intl ...
loby verify csv submit --file ... | status <id> | download <id>  # us_csv_verifications

loby identity create --first-name --last-name --address-line1 ... # identity_validation
loby identity get <id>

loby events list [--resource postcards] [--event-type letter.created]
loby events tail [--resource ...]                                # NDJSON stream

loby bank-accounts create --routing --account --signatory ...
loby bank-accounts verify <id> --amounts 11,35
loby billing-groups create --name --description

loby qr-codes create --redirect-url <url>
loby url-shortener create <long-url>

loby uploads create --campaign-id <id> --file <csv>
loby uploads status <id>
loby uploads errors <id>

loby resource-proofs get <id>

loby account
```

### Introspection (mandatory for agent discovery)

```
loby schema --json                       # full CLI tree
loby schema <command-path> --json        # one command's signature
loby exit-codes --json                   # the canonical exit code table
loby auth status --json                  # current profile, key prefix, environment
loby version --json                      # version, commit, build date, go version
loby completion bash|zsh|fish|powershell # shell completion script
```

### Exit codes

```
0   success
1   general error (permanent)
2   usage / argument error (permanent)
3   empty result (not an error — important for agents)
4   auth required / invalid key (permanent)
5   not found (permanent)
6   forbidden / insufficient permissions (permanent)
7   rate limited (transient — retry with backoff)
8   retryable error (transient — timeout, 5xx, network)
9   payment required / Lob domain-specific (permanent)
10  config error (permanent)
```

## Auth

**Precedence (highest first):**
1. `--api-key` flag
2. `LOB_API_KEY` env
3. Keyring entry for `--profile` (default: `default`)
4. Fail with exit code 4 and the line: `loby auth login` to fix

**Profiles:** `loby auth login --profile prod` stores the key under that name. `LOB_PROFILE` env selects the active profile. Test vs live is inferred from key prefix (`sk_test_` vs `sk_live_`) and surfaced in `auth status`.

**No secrets in config files. Ever.**

## Idempotency

Lob honors `Idempotency-Key` on POST and caches the response for 24h server-side. We:
1. Auto-generate `loby-<sha256(command+sorted-flags+body)[0:16]>` for every create call when the caller didn't pass `--idempotency-key`.
2. Send it as the `Idempotency-Key` header.
3. Inspect Lob's `Idempotent-Replayed` response header; surface it as `_replayed: true` in JSON output (still exit 0).

This makes the CLI safe for agents — they can call `loby postcards create ...` ten times during a flaky session and exactly one postcard mails. Cache lives at Lob; we don't reinvent it locally.

## SKILL.md

Lives at `skills/loby/SKILL.md` in the repo; users install with:
```bash
mkdir -p ~/.claude/skills/loby
curl -fsSL https://loby.voska.org/skill/SKILL.md > ~/.claude/skills/loby/SKILL.md
curl -fsSL https://loby.voska.org/skill/install.sh | sh   # full bundle
```

**Description field (focused on triggers, not workflow per Anthropic CSO guidance):**
> Use when sending physical mail (postcards, letters, checks, self-mailers, cards, booklets, buckslips, snap packs) via Lob, verifying US or international addresses, autocompleting addresses, looking up ZIP codes, managing direct-mail campaigns, working with mail templates, or interacting with any Lob API resource from the command line.

Structure: SKILL.md (under 200 words core, links to references/) + `references/COMMANDS.md` (full command catalog generated from `loby schema --json`) + `references/RECIPES.md` (verified mail flows: "send a postcard end-to-end", "verify and mail bulk", "manage a campaign").

Skill teaches the agent to:
1. Check `loby --version` (install via `brew install voska/tap/loby` if missing).
2. Set up auth via `loby auth status` then `loby auth login --api-key $LOB_API_KEY`.
3. Always pass `--json` and rely on documented exit codes.
4. Use `--dry-run` to preview mutating ops.
5. Discover unknown commands via `loby schema --json | jq`.

## Distribution & discovery

| Channel | Mechanism |
|---------|-----------|
| Homebrew (macOS/Linux) | `brew install voska/tap/loby` — goreleaser auto-PRs to `voska/homebrew-tap` |
| Scoop (Windows) | `scoop bucket add voska https://github.com/voska/scoop-bucket && scoop install loby` |
| Direct binary | GitHub Releases with `.tar.gz` + `.zip` + checksums + cosign signatures |
| go install | `go install github.com/voska/loby/cmd/loby@latest` |
| Landing page | `loby.voska.org` (GH Pages) with single-command install + llms.txt |
| Agent crawl | `loby.voska.org/llms.txt` enumerates pages; SKILL.md, COMMANDS.md, RECIPES.md, install.sh all served raw |
| pkg.go.dev | automatic |

`install.sh` autodetects OS/arch, downloads the right release, verifies checksum, drops the binary on PATH. The agent-discoverable one-liner:

```bash
curl -fsSL https://loby.voska.org/install.sh | sh
```

## Testing

- **Unit:** `*_test.go` next to source, race detector on, `httptest` for HTTP mocks
- **Integration:** `//go:build integration` build tag; `LOB_API_KEY=sk_test_...` required; CI runs on PRs from trusted contributors only
- **Schema snapshot:** `loby schema --json` golden file in `tests/schema.golden.json`; mismatch fails CI (forces semver discipline)
- **Live smoke:** `scripts/live-test.sh` — `loby addresses verify`, `loby postcards create --dry-run`, `loby account`

## CI/CD

- `ci.yml`: matrix on linux + macOS + windows; `make ci` (fmt-check + lint + test + build)
- `release.yml`: on `v*` tag, run goreleaser → builds + Homebrew formula PR + Scoop manifest + GitHub Release
- `pages.yml`: deploy `site/` to GitHub Pages on `main` push

## Dependencies (intentionally minimal)

| Package | Purpose |
|---------|---------|
| `github.com/alecthomas/kong` | CLI parser |
| _none_ | API client is hand-rolled `net/http`; spec types generated from `lob-api-public.yml` |
| `github.com/99designs/keyring` | OS keychain |
| `github.com/muesli/termenv` | Terminal colors |
| `github.com/BurntSushi/toml` | Config |
| stdlib `net/http`, `log/slog`, `testing` | Everything else |

No `viper`, no `cobra` (we use Kong), no third-party HTTP clients.

## Code quality bar

This ships under github.com/voska. Standards:
- `golangci-lint v2` strict (errcheck, errorlint, gosec, staticcheck, gofumpt, revive, wrapcheck, bodyclose)
- gofumpt formatting; goimports with `-local github.com/voska/loby`
- No package > 500 lines without justification
- One responsibility per package; sealed via `internal/`
- Zero TODOs in shipped code
- 15% unit-test coverage floor (CI gate); full coverage in Codecov
- Every exported symbol has a doc comment (revive enforces)
- No comment is allowed to say *what* the code does — only *why* when non-obvious

## Milestones

1. **M0 — Scaffold (this session, hopefully):** repo, Makefile, golangci config, AGENTS.md, SPEC.md (this doc symlinked), Kong skeleton, version embedding, output formatters, auth keyring, schema introspection, `loby version` / `loby schema` / `loby auth status` working
2. **M1 — Foundational verbs:** addresses (verify/autocomplete/zip/CRUD), accounts, events, bank_accounts. End-to-end recipe tested.
3. **M2 — Mail creation:** postcards, letters, checks, self_mailers, cards. Idempotency cache wired. `--dry-run` audited on all.
4. **M3 — Campaigns + templates + creatives + uploads.**
5. **M4 — Remaining resources:** booklets, buckslips, snap_packs, qr_codes, url_shortener, geo, identity_validation, billing_groups, resource_proofs, bulk verifications.
6. **M5 — Release:** goreleaser config, Homebrew tap, Scoop bucket, GH Pages site, llms.txt, SKILL.md polished, README, CHANGELOG. Cut `v1.0.0`.

## Open questions

None for V1. Decisions above are settled; any deviation requires updating this spec.
