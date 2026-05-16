# Changelog

All notable changes to `loby` are documented here. Format: [Keep a Changelog](https://keepachangelog.com); versioning: [SemVer](https://semver.org).

## [Unreleased]

## [0.1.3] — 2026-05-16

End-to-end live verification against `api.lob.com` test environment.
`scripts/live-smoke.sh` now exercises every CLI command that the test key
can reach (47 pass, 15 skip, 0 fail) and surfaced the following real bugs:

### Fixed
- Mailer response decoding crashed on `expected_delivery_date` and
  `send_date` when Lob returned bare `YYYY-MM-DD` (it documents these as
  full timestamps but returns date-only for test-created resources).
  Introduced `lob.Date`, a flexible unmarshaler that accepts RFC3339 or
  YYYY-MM-DD.
- `loby templates create` was sending `engine_type`, which Lob rejected
  with HTTP 422. The correct field is `engine` (`legacy` | `handlebars`).
- `loby identity verify` was sending `first_name`/`last_name`. Lob expects
  a single `recipient` (or `company`) plus a US address. Flags reshaped
  accordingly.
- `loby letters cancel` / `loby checks cancel` / `loby snap-packs cancel`
  used `POST /<res>/<id>/cancel` (404). Lob's actual cancel mechanism is
  `DELETE /<res>/<id>`. `execCancel` rewritten.
- `loby checks get` crashed because the `bank_account` field comes back
  as an object on retrieve, not a string ID. Typed as `any`.

### Removed (endpoints that don't exist on Lob's public API)
- `loby postcards cancel` — postcards enter the USPS pipeline immediately
  on create; no cancel endpoint.
- `loby self-mailers cancel` — same.
- `loby identity get` — identity validations aren't addressable by ID.

## [0.1.2] — 2026-05-16

### Fixed
- `loby qr-codes` was hitting `/qr_codes` (404). The real path is `/qr_code_analytics`. Lob does not expose create/get for QR codes — the `qr-codes` group is now list-only (with `--scanned`, `--limit`, `--offset` filters), since QR codes are minted by embedding Lob's snippet in mailer HTML rather than via the API.
- `loby short-urls` was hitting `/short_urls` (404). Lob's URL shortener lives at `/links` and `/domains`. Replaced `short-urls` with two new groups: `loby links` (create/get/list/delete) and `loby domains` (create/get/list/delete).
- `loby geo reverse <lat> <lng>` failed when longitude was negative because Kong parsed `-122.4194` as a flag. Switched to `--lat`/`--lng` flags, which also fixes the wrong path (`/reverse_geocode_lookups` → `/us_reverse_geocode_lookups`).
- `loby uploads list` was sending `limit`/`before`/`after` query params, which Lob rejects with HTTP 400. The endpoint only accepts `--campaign-id`. Also handle the bare-array response shape (Lob returns `[…]` here, not the usual envelope).

### Breaking
- `loby short-urls` is gone; use `loby links` instead.
- `loby qr-codes create` and `loby qr-codes get` are gone — these endpoints do not exist on Lob's public API.
- `loby geo reverse` now takes `--lat` and `--lng` flags instead of two positional args.

## [0.1.1] — 2026-05-16

### Fixed
- `loby zip <code>` was issuing `GET /us_zip_lookups/:zip` and getting 404. Lob's actual endpoint is `POST /us_zip_lookups` with the zip in the body.
- `loby account` was hitting `GET /accounts` (401). The documented endpoint is `GET /accounts/credits_balance`.

Both bugs surfaced from end-to-end live testing against Lob's test environment.

## [0.1.0] — 2026-05-16

### Added
- Initial public release.
- Coverage for all 29 Lob v1 resources: addresses, postcards, letters, checks, self-mailers, snap packs, cards, booklets, buckslips, campaigns (+ informed delivery as multipart), creatives, uploads (+ exports, report), templates (+ versions), bank accounts, billing groups, QR codes, short URLs, events (list/tail/get), resource proofs, accounts, identity validation, ZIP lookup, reverse geocode, US/intl verifications (single + sync bulk), US autocompletion.
- Output modes: human, `--json`, `--plain`, `--select` field projection, `--results-only` envelope strip, `--quiet` bare-IDs, NDJSON streaming for `events tail`.
- Auth via OS keychain (99designs/keyring) + `LOB_API_KEY` env + named profiles; `loby auth login` uses `term.ReadPassword` so secrets never echo into scrollback.
- HTTP client with HTTP Basic auth, deterministic auto-generated `Idempotency-Key`, rate-limit-aware retries honoring both `Retry-After` and `X-Rate-Limit-Reset`, structured error envelope, replay detection, 100MB multipart cap.
- Introspection: `loby schema --json`, `loby exit-codes --json`, `loby version --json`, `loby auth status --json`, `loby completion bash|zsh|fish|powershell`.
- Safety: `--dry-run` on every mutation, `--confirm` required for destructive ops, agent input validation (paths/IDs/control chars), `--no-input` auto-implied on non-TTY stdin.
- Binary file inputs (PDF, PNG, JPG, etc.) auto-encoded as `data:` URIs; text inputs (HTML, CSV, MD) pass through inline.
- Distribution: Homebrew tap (`voska/tap`), Scoop bucket (`voska/scoop-bucket`), signed binaries with cosign keyless OIDC, SBOM via Syft, build provenance attestation.
- CI/CD: GitHub Actions on Linux/macOS/Windows × amd64/arm64, golangci-lint v2, 15% unit-test coverage floor (plus Codecov tracking), schema-snapshot regression guard, CodeQL scan, Dependabot for go modules + GHA, GH Pages deploy at <https://lobycli.com>.
- Canonical [SKILL.md](skills/loby/SKILL.md) for AI agents with [command catalog](skills/loby/references/COMMANDS.md), [verified recipes](skills/loby/references/RECIPES.md), and [resource glossary](skills/loby/references/RESOURCES.md).
- Custom domain <https://lobycli.com> with `install.sh`, `llms.txt`, and the full SKILL bundle for agent discovery.

[Unreleased]: https://github.com/voska/loby/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/voska/loby/releases/tag/v0.1.3
[0.1.2]: https://github.com/voska/loby/releases/tag/v0.1.2
[0.1.1]: https://github.com/voska/loby/releases/tag/v0.1.1
[0.1.0]: https://github.com/voska/loby/releases/tag/v0.1.0
