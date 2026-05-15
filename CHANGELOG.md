# Changelog

All notable changes to `loby` are documented here. Format: [Keep a Changelog](https://keepachangelog.com); versioning: [SemVer](https://semver.org).

## [Unreleased]

### Added
- Initial public release.
- Coverage for all 29 Lob v1 resources: addresses, postcards, letters, checks, self-mailers, snap packs, cards, booklets, buckslips, campaigns (regular + informed delivery), creatives, uploads, templates (+ versions), bank accounts, billing groups, QR codes, short URLs, events (list/tail/get), resource proofs, accounts, identity validation, ZIP lookup, reverse geocode, US/intl verifications (sync, bulk, CSV), US autocompletion.
- Output modes: human, `--json`, `--plain`, `--select` field projection, NDJSON streaming for `events tail`.
- Auth via OS keychain (99designs/keyring) + `LOB_API_KEY` env + named profiles.
- HTTP client with HTTP Basic auth, auto-generated `Idempotency-Key`, rate-limit-aware retries with `Retry-After`, structured error envelope, replay detection.
- Introspection: `loby schema --json`, `loby exit-codes --json`, `loby version --json`, `loby auth status --json`.
- Safety: `--dry-run` on every mutation, `--confirm` required for destructive ops, agent input validation (paths/IDs).
- Distribution: Homebrew tap (`voska/tap`), Scoop bucket (`voska/scoop-bucket`), signed binaries with cosign, SBOM via Syft, build provenance attestation.
- CI/CD: GitHub Actions on Linux/macOS/Windows, golangci-lint v2, 80% coverage gate, schema snapshot test, CodeQL scan, Dependabot for go modules and GHA, Pages deploy.
- Canonical [SKILL.md](skills/loby/SKILL.md) for AI agents with [command catalog](skills/loby/references/COMMANDS.md), [verified recipes](skills/loby/references/RECIPES.md), and [resource glossary](skills/loby/references/RESOURCES.md).

[Unreleased]: https://github.com/voska/loby/compare/main...HEAD
