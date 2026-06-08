# Fugo — agent guide

**v0.1.0** · Go SDUI framework. Infrastructure layer only (CI, linting, repo skeleton). Zero application code exists yet — no SDK, no Flutter client, no CLI, no transport layer.

## Commands

```sh
make test         # go test ./... -count=1 -race -shuffle=on -v
make lint         # golangci-lint run --timeout 10m ./...
make vet          # go vet ./...
make build        # builds bin/fugo with version/commit/date ldflags
make install      # go tool lefthook install (git hooks)
make release TYPE=patch MSG="description"   # bumps VERSION, updates CHANGELOG, commits, tags
make pr MSG="type: description"             # creates branch, pushes, opens PR via gh
make pr-merge PR=5                          # squash-merges and cleans up
```

## Git hooks (Lefthook v2.1.9)

**Pre-commit** (parallel, auto-fixes staged files): golangci-lint --fast --fix → go vet → gofumpt → go mod tidy.

**Pre-push** (sequential, full suite): golangci-lint → go vet → staticcheck → go build → go test -race → go mod tidy.

Pre-push also verifies VERSION and CHANGELOG.md are in sync.

## CI (push/PR to `main`)

Five parallel jobs: golangci-lint, staticcheck, go vet, go build, go test -race -shuffle=on, gofumpt formatting check.

## Conventions

- **Formatter**: `gofumpt` (not `gofmt`). Run `gofumpt -w .` before committing.
- **Linters**: 80+ linters enabled via `.golangci.yml`. `staticcheck` also enforced.
- **PR checklist** (from template): gofumpt, go vet, golangci-lint, go test -race, go mod tidy, CHANGELOG updated.
- **Release process**: Use `make release`, never manually edit VERSION or CHANGELOG.

## Repo quirks

- `README.md` is empty. `SPEC.md` is the real specification. `ROADMAP/` is in Spanish.
- `doc.go` is the only Go source file — `package fugo`.
- Module path: `github.com/sazardev/fugo`, Go 1.26.3.
- Agent skills installed: `caveman`, `caveman-commit`, `golang-patterns`, `golang-testing` (`.agents/skills/`).
- All dependencies are currently `// indirect` (pulled by lefthook tool dependency).
