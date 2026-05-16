# loby — repo build contract for AI agents

`loby` is a CLI for Lob (direct mail). Data goes to stdout (parseable). Progress, hints, and errors go to stderr. Never mix.

## Build & test

```
make build           # bin/loby
make test            # unit, race detector on
make test-integration # requires LOB_API_KEY=sk_test_...
make lint            # golangci-lint v2
make fmt             # gofumpt + goimports
make ci              # fmt-check + vet + lint + test + build (the gate)
```

## Layout

```
cmd/loby/         Thin main.go: version embed + Kong dispatch.
internal/cli/     Kong command structs. One file per resource group.
internal/client/  Hand-rolled HTTP client over Lob's API. Auth, retries, idempotency.
internal/lob/     Per-resource types and methods. One file per Lob resource.
internal/output/  Output formatters (human, json, plain, ndjson) + --select projection.
internal/auth/    OS keychain (99designs/keyring) + LOB_API_KEY env + named profiles.
internal/config/  XDG config (TOML). State (idempotency cache) lives under ~/.config/loby/state/.
internal/errfmt/  Maps API errors → exit codes (0/1/2/3/4/5/6/7/8/9/10) + recovery hints.
internal/schema/  Walks Kong AST to emit `loby schema --json`.
internal/version/ Build info embedded via ldflags.
skills/loby/      Canonical agent skill (drop into ~/.claude/skills/loby/).
site/             GitHub Pages: landing + llms.txt + install.sh.
```

## Output modes

Every command supports: default (human/colored TTY), `--json/-j`, `--plain/-p` (TSV), `--results-only`, `--select f1,f2.nested`. NDJSON for list/tail/streaming. Stdout is parseable; stderr is human.

## Exit codes

`0` success · `1` error · `2` usage · `3` empty · `4` auth · `5` not_found · `6` forbidden · `7` rate_limited · `8` retryable · `9` payment_required · `10` config_error.
Authoritative source: `loby exit-codes --json`.

## CLI conventions

- Flags are kebab-case (`--mailing-date`, `--bank-account`).
- Every mutating verb supports `--dry-run/-n` (returns the would-be request body as JSON and exits 0).
- Every mutating verb auto-generates an idempotency key unless `--idempotency-key` is set. Keys are persisted to `~/.config/loby/state/idempotency.sqlite`.
- `--no-input` makes the CLI fail rather than prompt. Auto-detected when stdin is not a TTY.
- `--confirm` or `--force` required for `delete`.
- `LOBY_AUTO_JSON=1` flips default output to JSON when stdout is piped.
- Auth precedence: `--api-key` > `LOB_API_KEY` env > keyring (`--profile`, default `default`).

## Introspection (every CLI MUST have)

```
loby schema --json                # full command tree, all flags
loby schema <cmd> --json          # one command's signature
loby exit-codes --json            # exit code table
loby auth status --json           # active profile, key prefix, environment
loby version --json               # version, commit, date, go runtime
```

## Code quality

- `make ci` is the gate. CI fails on any lint or test failure.
- gofumpt formatting. goimports `-local github.com/voska/loby`.
- No package > 500 LOC without justification. No file > 300 LOC ideally.
- One responsibility per package. Everything not exported lives under `internal/`.
- Comments explain `why`, never `what`. No comment may reference removed code, the current task, or callers.
- Zero TODOs in shipped code.
- Every exported symbol has a doc comment (revive enforces).
- Unit-test coverage floor: 15% (CI gate guards against core regressions; full coverage tracked in Codecov).

## Commits & PRs

Conventional Commits: `feat(scope):`, `fix(scope):`, `chore:`, `docs:`, `test:`, `refactor:`. Imperative present-tense subject under 70 chars. PR body bullets the why, not the what. Update `CHANGELOG.md` under `## Unreleased`.

## Security

- Never commit API keys, `.env`, or credential files. (`.gitignore` enforces.)
- Keys live in OS keychain via 99designs/keyring. CI uses `LOB_API_KEY` env from secrets.
- No third-party HTTP clients. stdlib `net/http` only.
- Validate every agent-supplied path/ID/flag (canonicalize paths, reject control chars, reject `%?#` in resource IDs).

## Release

`git tag v$X.Y.Z && git push --tags` triggers `release.yml` → goreleaser → multi-platform binaries + Homebrew tap PR + Scoop manifest + GitHub Release with cosign signatures. Update `CHANGELOG.md` first.
