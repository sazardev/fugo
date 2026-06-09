# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-06-09

### Added
- Native **Material 3** rendering, **light by default**. The Flutter client builds its `ColorScheme` from the active `fg.Theme` (forwarded as `FUGO_THEME_SEED` + `FUGO_THEME_BRIGHTNESS`) via `ColorScheme.fromSeed`.
- Material button variants as separate constructors: `fg.FilledButton`, `fg.FilledTonalButton`, `fg.OutlinedButton`, `fg.TextButton`, `fg.ElevatedButton`, `fg.IconButton` (`fg.Button` is an alias of `FilledButton`). Buttons gain a leading `.Icon()` and `.Enabled()`.
- Core Material widgets: `fg.Card`, `fg.Scaffold` (`.AppBar` / `.FAB`), `fg.FloatingActionButton`, `fg.ListTile`, `fg.Chip`, and `fg.ProgressCircular` / `fg.ProgressLinear`.
- `fg.Column` alignment controls: `.MainAlign`, `.CrossAlign`, `.MainAxisSize`, `.Expand`.

### Changed
- The renderer **auto-centers** an intrinsically-sized root (e.g. a bare `Column`), so simple content lands in the middle of the window without an explicit `Center`. Roots that fill the viewport (`Scaffold`, `Container`, `ListView`, …) are left as-is.
- Widgets no longer inject opinionated hex colors: `Text`, `Container`, and `Button` inherit the Material 3 `ColorScheme` unless a color setter is called. The default theme is now `LightTheme`.
- `fugo init`'s counter template is now minimal (no manual colors or padding), relying on the Material 3 defaults and auto-centering.

## [0.3.2] - 2026-06-09

### Added
- `LICENSE` file (MIT). The repository had no license file, so the module was non-redistributable and **pkg.go.dev hid the rendered documentation** ("License: None detected"). Adding the standard MIT text restores documentation on pkg.go.dev and matches the README's stated license.

## [0.3.1] - 2026-06-09

### Fixed
- `fugo --version` now reports the correct version when the CLI is installed with `go install` (the Makefile's `-ldflags` are not applied in that path, so it previously printed the hardcoded `0.1.0`). It falls back to the module version and VCS stamps from `runtime/debug.ReadBuildInfo()`.

## [0.3.0] - 2026-06-09

### Added
- **Installable via `go install`**: `go install github.com/sazardev/fugo/cmd/fugo@latest` now works. The generated protobuf Go bindings are committed, so a clean module-proxy fetch (or fresh clone) builds the CLI without `protoc` or any code-gen step. Added a README **Installation** section documenting the CLI install and the Flutter rendering prerequisite.
- OS host services: clipboard (`Context.Clipboard()`) and native file dialogs (`Context.Files().Open/Save`), answered asynchronously by the Flutter client and dispatched on the event goroutine.
- Runtime window control via `Context.Window()` (`window_manager`-backed): set title/size, minimize, maximize, center, fullscreen.
- New widgets: `fg.AnimatedPositioned` (animate a child between positions inside a `Stack`) and `fg.WindowDragArea` (make a region drag a frameless window).
- Scheduler immediate-priority path: `Context.UpdateNow()` wakes the render loop without waiting for the next 16ms tick.

### Changed
- Generated Go protobuf bindings (`transport/proto/fugo/v1/*.pb.go`) are **no longer gitignored** — they are committed and kept gofumpt-clean by `make proto` (which now runs `gofumpt -w` on the generated Go). The CI `format` job validates the committed bindings; lint/vet/build/test/bench still regenerate them via the `gen-proto` action to catch proto drift.
- Performance: object-pooled diff lookup map, GC tuning (`FUGO_GOGC` / `FUGO_GOMEMLIMIT`), and Go + Dart benchmarks behind a CI perf-regression gate.

## [0.2.0] - 2026-06-08

### Added
- CLI DX overhaul: leveled verbose and quiet logging, animated steps, rich help, a widgets catalog, and a FUGO_LOG runtime logger

### Fixed
- Data race on `App.handlers` between the scheduler and transport goroutines (now mutex-guarded).
- Keyed diff desync: cross-frame `Key` matching emitted patches for node ids the client never received; node identity is now positional by id, so every patch is client-applicable.
- Text (font weight, alignment) and Container (per-edge padding) properties that were settable but never reached the wire.

### Changed
- Corrected the `CLAUDE.md` `ui`↔`fg` / `NewText` inversion and added FlatBuffers→Protobuf correction banners across the ROADMAP.
- Re-baselined performance budgets on standard protobuf with a deterministic zero-alloc diff gate; CI now regenerates the gitignored Go protobuf bindings via a composite action.
- Promoted `golang.org/x/term` and `golang.org/x/sys` to direct dependencies.

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
