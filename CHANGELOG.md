# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-07

### Added
- Root package declaration (`doc.go`)
- Lefthook git hooks config with `assert_lefthook_installed`
- Pre-commit hooks: golangci-lint (fast), go vet, gofumpt, go mod tidy
- Pre-push hooks: golangci-lint (full), go vet, staticcheck, go build, go test, go mod tidy
- GitHub Actions CI with lint, vet, build, test, format jobs
- `.golangci.yml` with 80+ linters (govet, staticcheck, nilnil, gocritic, revive, gosec, etc.)
- `.gitattributes` for cross-platform line endings
- `Makefile` for test/build/lint/release workflows
- `VERSION` file for semver tracking
- `CHANGELOG.md` with Keep a Changelog format

### Changed
- Replaced deprecated `tenv` linter with `usetesting`
- Switched golangci-lint CI installation from action binary to `go install` for Go 1.26 compat

### Infrastructure
- Go module initialized at `github.com/sazardev/fugo` (Go 1.26.3)
- Lefthook v2.1.9 as tool dependency
- Agent skills for caveman and Go patterns/testing
