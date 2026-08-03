# Contributing to Fugo

Fugo is a local Server-Driven UI framework: you write everything in Go, a precompiled
Flutter client renders it. Contributions to either side — the Go engine or the Dart
render client — are welcome.

## Before you start

- Read `CLAUDE.md` at the repo root first — it's the canonical architecture doc (packages,
  data flow, wire format, conventions) and stays current with the code, unlike some of the
  older Spanish-language design docs under `ROADMAP/`.
- For anything non-trivial (a new widget, a transport change, a CLI command), open an issue
  or discussion before writing code — it saves rework if the approach needs adjusting.
- Small fixes (typos, docs, a clear bug fix with a regression test) can go straight to a PR.

## Development setup

Requires **Go 1.26+**. The generated protobuf bindings are committed, so `go build ./...`
works out of the box without `protoc`.

```bash
git clone https://github.com/sazardev/fugo && cd fugo
make install        # installs Lefthook git hooks (auto lint/format/test on commit & push)
go build ./...
make test            # go test ./... -count=1 -race -shuffle=on -v
```

If you're changing `transport/proto/fugo/v1/fugo.proto`, you'll additionally need `protoc`,
`protoc-gen-go`, `protoc-gen-go-grpc`, and `protoc-gen-dart` (see `make proto` and the
version-pin note in `CLAUDE.md` — the Dart plugin **must** be `protoc_plugin` 21.1.x).

If you're touching `flutter_client/` (the Dart render client), you'll need the
[Flutter SDK](https://docs.flutter.dev/get-started/install) pinned in `FLUTTER_VERSION` —
`fugo doctor` warns if your installed SDK doesn't match. See `flutter_client/README.md`.

## Code style

- **Formatter is `gofumpt`**, not `gofmt` — run `gofumpt -w .` before committing (Lefthook
  does this for staged files automatically).
- `make lint` runs `golangci-lint` (80+ linters configured in `.golangci.yml`) and
  `staticcheck`. Both must pass clean — there's no threshold, every issue counts.
- `make vet` (`go vet ./...`) must pass.
- Keep changes scoped: a bug fix shouldn't carry an unrelated refactor. No speculative
  abstractions for hypothetical future needs.
- Widgets follow the `fg/` conventions: prefix-free constructors (`fg.Text`, not
  `NewText`), chainable setters, a `*Set bool` flag for optional props so proto3 can omit
  an unset field cleanly. Adding a widget touches **four places** — see `CLAUDE.md` →
  "Wire format".

## Tests

- `make test` runs the full suite with `-race -shuffle=on`. New code needs test coverage;
  bug fixes need a regression test that fails before the fix and passes after.
- Table-driven tests are the house style for the Go side (see any `*_test.go` for the
  pattern already in use).
- Benchmarks live in `*_bench_test.go` (`engine/` has the diffing perf gates CI checks on
  every push — don't regress them without a good reason, documented in the PR).

## Commit & PR flow

1. Branch off `main`.
2. Make your change, keep commits focused.
3. `gofumpt -w .`, `make lint`, `make vet`, `make test`, `go mod tidy` — all clean.
4. Update `CHANGELOG.md` under `[Unreleased]` if the change is user-facing (the pre-push
   hook checks `VERSION` has a matching entry on release, but documenting as you go makes
   that step painless).
5. Open the PR — `.github/pull_request_template.md` will prompt you for the checklist above.
   `make pr MSG="type: description"` does steps 2-5's branch/push/PR-open plumbing if you're
   working from `main` directly.

CI runs lint, vet, build, test (`-race -shuffle=on`), a `gofumpt` format check, and the perf
gate/benchmarks on every push and PR — all must pass before merge. CodeQL runs weekly and on
every push for security analysis.

## Releases

Don't hand-edit `VERSION`, `FLUTTER_VERSION`, or `CHANGELOG.md` — those are maintainer-run via
`make release` / `make release-flutter` (see `CLAUDE.md` for the versioning scheme, which
tracks the exact Flutter release Fugo renders against).

## Reporting bugs / requesting features

Use the issue templates — they ask for the right context up front (repro steps, Go/Flutter
versions, expected vs actual behavior). For security issues, see `SECURITY.md` instead of
opening a public issue.

## Code of Conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md). By participating, you're
expected to uphold it.
