# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **H1 "Hello Fugo" spike**: Go sends widget tree to Flutter via gRPC over TCP localhost
- **H2 "Counter App"**: Functional Counter demo with full event loop (click → event → state change → diff → re-render)
- `transport/proto/fugo/v1/fugo.proto`: service + messages (RenderPayload oneof full_tree|patches, WidgetTree, WidgetNode, PatchList, Patch, TextProps, ButtonProps, ContainerProps, ClientEvent)
- `transport/server.go`: gRPC server with lazy stream binding, pending buffer, AppHandler interface
- `ui/widget.go`: Widget interface, baseWidget, BuildTree/collectIDs
- `ui/text.go`: Text widget with SetText mutation
- `ui/button.go`: Button widget with OnClick handler
- `ui/container.go`: Container widget with bg_color/padding/border_radius props
- `ui/layout.go`: Column and Center widgets with recursive tree walking
- `engine/differ.go`: O(n) diff algorithm (CREATE, UPDATE, DELETE, REPLACE, REORDER) with 7 unit tests
- `engine/reconciler.go`: Lazy stream binding with pending message buffer, SendFullTree/SendPatches
- `app.go`: App (Run, Shutdown, HandleEvent), Context (Update), state loop with diff/patch cycle
- `supervisor/process.go`: Flutter subprocess lifecycle management (spawn, signals, shutdown)
- `cmd/fugo-spike/main.go`: Counter demo entry point
- `flutter_client/`: Flutter project with gRPC client, widget registry (Text, Button, Container, Column, Center), patch application, event emission
- Makefile targets: `proto`, `flutter-build`, `spike`, `run-spike`

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
