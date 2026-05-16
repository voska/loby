# Changelog

All notable changes to `loby` are documented here. Format: [Keep a Changelog](https://keepachangelog.com); versioning: [SemVer](https://semver.org).

## [Unreleased]

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

[Unreleased]: https://github.com/voska/loby/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/voska/loby/releases/tag/v0.1.1
[0.1.0]: https://github.com/voska/loby/releases/tag/v0.1.0
