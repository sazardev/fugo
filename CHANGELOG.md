# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `Context.Notifications().Show(title, body)` — native OS notifications (toast / notification center), backed by `local_notifier` on the client.
- `DataTable` gained `.Sortable(handler)`/`.SetSort` (tappable, sortable column headers) and `.Selectable(handler)`/`.SetSelected` (a leading checkbox column) — Go still owns reordering/selection state.
- `fg.Form(items...)` — aggregates per-field validation via `.AddField(validate func() string)` + `.Validate() bool`, plus a form-level `.SetError` banner; a plain layout wrapper, since Fugo has no client-owned FormField state to orchestrate.
- `fg.Dismissible(child)` — swipe-to-dismiss, with `.Direction`, `.BgColor`, `.Icon`, and `.OnDismissed(handler)`.
- `fg.SliverScaffold(title, body...)` — a scrollable screen with a collapsing `SliverAppBar` (`.ExpandedHeight`, `.Pinned`, `.Floating`, `.FlexibleBackground`) over a `SliverList` body.

- `fg.Store[S]` — an opt-in, generic, mutex-protected piece of shared state (`Get`/`Update(fn)`/`Subscribe`) for apps where several widgets react to the same data; composes with `Context`/`Update()`/the retained tree rather than replacing them.
- `Context.RequestFocus(w fg.Focusable)` — asks the client to move keyboard focus to a widget (today, `TextField`); useful to jump back to a field after a validation error.
- `fg.Responsive(base).At(minWidth, child)` — picks one of several prebuilt children based on the actual available width, decided entirely client-side via a `LayoutBuilder` (no round trip to Go).
- `fg.Draggable(child).Data(...)` / `fg.DragTarget(child).OnAccept(handler)` — drag-and-drop between widgets.
- `Context.OnFileDrop(func(paths []string))` — reports files dragged from the OS onto the client window, backed by `desktop_drop`.
- `Theme.FollowSystem` — tracks the OS's light/dark setting live (wires up the previously-orphaned `FUGO_THEME_FOLLOW_SYSTEM` the client already read). `Theme.Typography.Family` is now actually forwarded to the client (`FUGO_THEME_FONT_FAMILY`) and applied via `ThemeData.fontFamily` (a system-installed font by name — Fugo doesn't bundle font assets).
- `AppBar` gained `.Elevation`, `.ForegroundColor`, and `.Bottom(widget)` (a fixed-height strip docked under the title row, e.g. a search field or filter chips).
- `fugo autostart enable`/`disable` — registers the built `dist/` app to launch at login (XDG autostart `.desktop` entry on Linux, a `Run` registry key on Windows).
- `fugo build` now also writes Linux packaging next to the binary: an XDG `.desktop` entry, a per-user `install.sh` (no root needed), and an AUR `PKGBUILD` template.
- `fugo doctor` checks for `notify-send`/libnotify on Linux (used by `Context.Notifications`), as a non-blocking warning.

### Fixed
- `TextField`/`Dropdown` actually wire up `.SetError` now (the proto's `error_text` field existed and the Flutter client already rendered it, but the Go widgets never exposed a setter or marshaled it).
- Native file dialogs (`Context.Files().Open`/`.Save`) actually work on Linux now. The client used `file_picker`, which has no Linux platform implementation at all (only android/ios/web/macos/windows) — it built fine but threw at runtime the moment a Linux user opened a dialog. Replaced with `file_selector` (Flutter-official, federated), which has real native implementations for Linux, Windows, and macOS alike.

## [3.44.8-fugo.0] - 2026-08-02

### Added
- 13 new widgets: `CheckboxListTile`, `RadioListTile`, `SwitchListTile`, `NavigationRail`, `PageView`, `Table`, `ConstrainedBox`, `FractionallySizedBox`, `VerticalDivider`, `RangeSlider`, `Autocomplete`, `Scrollbar`, `RefreshIndicator`.
- `fg.Semantics(child)` — accessibility annotations (label/hint/button/header) for screen readers.
- `fg.Canvas(w, h)` with `.Line`/`.Rect`/`.Circle`/`.Path`/`.FilledPath` — a bounded, declarative drawing surface for sparklines/gauges/simple charts.
- `Context.RegisterShortcuts(map[string]func())` — app-wide keyboard shortcuts (e.g. `"ctrl+s"`), matched client-side independent of focus.
- `Context.OnResize(func(width, height float64))` — the one channel back for Go to learn the client's viewport size.
- `GestureDetector` gained `OnDoubleTap`, `OnLongPress`, `OnPanStart/Update/End`, `OnScaleStart/Update/End`.
- `Router` gained `.OnBeforeLeave(func() bool)` (veto a navigation, e.g. "discard unsaved changes?") and now animates route transitions client-side; it also has its own `WidgetType` instead of reusing `CONTAINER`.
- `TextField`/`Dropdown` gained `.SetError(msg)` for Go-owned validation state.
- `Context.ShowDialogActions`/`ShowBottomSheetActions` — dialogs/bottom sheets with custom action buttons (not just a single "OK"), reusing the same reply channel as the date/time pickers.
- `Container` gained `.Margin`, `.Border`, `.Shadow`; `Text` gained `.MaxLines`, `.Overflow`, `.Decoration`, `.LetterSpacing`, `.Italic`; `TextField` gained `.MaxLines`, `.PrefixIcon`, `.SuffixIcon`, `.KeyboardType`.
- `Theme.Components` (`CardRadius`, `CardElevation`, `ButtonRadius`) — per-component theme defaults, Flutter's `CardTheme`/`ElevatedButtonThemeData` equivalent.
- `fg` testing kit (`ClickEvent`, `BoolEvent`, `TextEvent`, `FloatEvent`, `RangeEvent`) and `BuildTree` documented as the supported way to test a Fugo UI without gRPC/Flutter.
- `FUGO_THEME_FOLLOW_SYSTEM=1` makes the client track the OS light/dark setting live via `ThemeMode.system`.
- The Flutter client now maps ~30 `Curves` (was 4), warning on an unrecognized name instead of silently falling back to `ease`.

## [0.17.0] - 2026-06-09

### Added
- `fugo doctor` now checks **project coherence**: it reads `go.mod`'s module path and `main.go`'s `<module>/ui` import and reports a precise ✗ when they don't match, when the imported `ui` package is missing, or when it has no exported `Build` — caught before the generic compile error.
- `fugo doctor --fix` repairs the auto-fixable bits inside a project before diagnosing: writes a default `fugo.toml` if missing, runs `git init` if there's no repo, and `go mod tidy` to resolve dependencies.

### Changed
- **`fugo run` hot-reloads by default.** It now watches `.go` files and rebuilds/reconnects the Go server on every change while the Flutter window stays open — so edits to text, handlers, layout, etc. show up live (in-memory state resets across reloads). Pass `--no-watch` for a single build-and-run; the old `--watch` flag is now implied (hidden, no-op).

### Fixed
- Flutter render client: reconnection after a server restart (hot reload) no longer fails with "Bad state: Stream has already been listened to". The isolate now subscribes to its `ReceivePort` once and routes events to the active stream, and resets its backoff after a successful session so it reconnects within ~500ms.

## [0.16.0] - 2026-06-09

### Added
- `fugo doctor` is now a two-part health check. **Toolchain**: Go, Flutter, git, protoc, gofumpt (Go/Flutter are required → ✗; the rest are warnings) plus the platform. **Project** (when run inside a Fugo project): it imports `fugo.toml` and reports the resolved window/address, checks the configured gRPC address is free (or a unix socket), verifies the structure (`main.go`, `ui/`, `go.mod`), that the `github.com/sazardev/fugo` module resolves, whether the Flutter client is built (or buildable), and finally compiles the project (`go build ./...`).

### Changed
- `fugo doctor` now exits non-zero when there's a blocking issue (a required tool missing, an unresolved module, or a project that doesn't compile), so it can gate scripts/CI. Warnings alone keep a zero exit.

## [0.15.0] - 2026-06-09

### Added
- `fugo init` now scaffolds a recommended project structure instead of a lone `main.go`: a thin `main.go` entrypoint, a `ui` package for your screens (`ui.Build`), a `fugo.toml` config, a `README.md`, a `.gitignore`, and `logs/` (with `bin/`/`dist/` reserved for build output). It also runs `git init` and makes an initial commit (skip with `--no-git`).
- `fugo.toml` — declarative project config read by the CLI **and** the app: `[window] title/width/height` and `[server] addr`. Generated `main.go` loads it via the new `fugo.ConfigOptions("fugo.toml")`; `fugo run` uses `[server] addr` as the default address (still overridable with `--addr`), and `fugo build` ships `fugo.toml` into `dist/`.
- New dependency-free `config` package (`github.com/sazardev/fugo/config`) that loads `fugo.toml` (a small, fixed TOML subset), shared by the runtime and the CLI.
- `fugo run` now tees the app's runtime logs to `logs/run.log` (still streamed to the console).

### Changed
- The `app` and `showcase` starter templates now live in the generated `ui` package (`ui.Build`); the theme is set in `main.go` before `RunStandalone`.

## [0.14.0] - 2026-06-09

### Added
- `fg.RichText(fg.Span("a").Bold(), fg.Span("b").Color(...).Size(...))` — a paragraph of mixed-style text runs.
- `fg.DataTable().Columns(...).Row(...)` — a Material data table (horizontally scrollable).
- `fg.Stepper().Step(title, content).Active(i).OnStep(fn)` — a step-by-step wizard (the tapped step index arrives via OnStep). (Pedido A continuation — wave 8; rounds out the planned Material catalog.)

## [0.13.0] - 2026-06-09

### Added
- More imperative overlays from Go: `ctx.ShowBottomSheet(title, message)` (modal bottom sheet) and the native pickers `ctx.PickDate(func(date string))` / `ctx.PickTime(func(t string))`, which return the chosen value (ISO `YYYY-MM-DD` / 24-hour `HH:MM`) to a callback — empty if cancelled. Pickers reuse the request/reply correlation of the host-service channel. (Pedido A continuation — wave 7.)

## [0.12.0] - 2026-06-09

### Added
- Layout helpers: `fg.AspectRatio(ratio, child)`, `fg.ClipRRect(radius, child)`, `fg.FittedBox(child)`, and `fg.Flexible(child).Flex(n)` (loose-fit counterpart of `Expanded`).
- `fg.ExpansionTile(title)` — a collapsible accordion (`.Subtitle`, `.Leading`, `.Children`, `.InitiallyExpanded`), expanded/collapsed on the client.
- `fg.PopupMenuButton(icon)` — an overflow/context menu (`.Item(value, label)`, `.OnSelected` with the chosen value in the event data).

## [0.11.0] - 2026-06-09

### Added
- More common Material widgets: `fg.Tooltip(message, child)`, `fg.Badge(child).Label(...)` (a dot when empty), `fg.CircleAvatar` (`.Text` / `.Icon` / `.BgColor` / `.Radius`), and `fg.SegmentedButton` (`.Item(value, label)`, `.Selected`, `.OnChange` with the selected value in the event data). (`fg.Spacer` already shipped as sugar over `Expanded`.)

## [0.10.0] - 2026-06-09

### Added
- Imperative overlays driven from Go over the out-of-band command channel (alongside window/host services): `ctx.ShowSnackBar(text)` shows a Material snackbar and `ctx.ShowDialog(title, message)` shows an alert dialog (dismissed with OK). (Pedido A wave 4 — completes the navigation/overlay series: AppBar, Drawer/NavigationBar, Tabs, Dialog/SnackBar.)

## [0.9.0] - 2026-06-09

### Added
- `fg.Tabs` — a Material tab strip with one view per tab (`.Tab(label, content)`, `.InitialIndex`). Tab switching is handled on the client, so it needs no round-trip to Go. (Pedido A wave 3.)

## [0.8.0] - 2026-06-09

### Added
- `Scaffold.Drawer(widget)` — a slide-in side panel (e.g. a `Column` of `ListTile`s). With an app bar present and no explicit leading, the menu button that opens it appears automatically.
- `fg.NavigationBar` — a Material 3 bottom navigation bar (`.Item(icon, label)`, `.Selected(i)`, `.OnChange` with the selected index in the event data), attached via `Scaffold.BottomBar(...)`. (Pedido A wave 2: navigation/layout.)

## [0.7.0] - 2026-06-09

### Added
- `fg.AppBar` — a full Material app bar: a title plus an optional `.Leading` widget (e.g. a menu icon button), trailing `.Actions(...)`, `.CenterTitle`, and `.BgColor`. (First of the navigation/layout wave: AppBar → Drawer/NavigationBar → Tabs → Dialog/SnackBar.)

### Changed
- **Breaking:** `Scaffold.AppBar` now takes an `*fg.AppBar` widget instead of a title string, matching Flutter's `Scaffold(appBar: AppBar(...))`. Migrate `fg.Scaffold(body).AppBar("X")` → `fg.Scaffold(body).AppBar(fg.AppBar("X"))`.

## [0.6.0] - 2026-06-09

### Added
- Flutter-style constant banks — stop hand-writing strings, hex and magic sizes:
  - **`fg.Icons.*`** — the full Material icon set (~2,200 base icons, e.g. `fg.Icons.Home`, `fg.Icons.Coffee`). Works with every icon-taking API (`fg.Icon`, `fg.IconButton`, `fg.FloatingActionButton`, `fg.ListTile`).
  - **`fg.Colors.*`** — the Material palette (`fg.Colors.Amber`, `.Blue`, `.RedAccent`, `.Grey800`, `.Transparent`, …).
  - **`fg.TextSize.*`** — the Material 3 type scale (`fg.TextSize.DisplayLarge` … `.LabelSmall`).
- `cmd/gen-icons`: a dev tool that regenerates `fg/icons_gen.go` and `flutter_client/lib/icons_gen.dart` from the installed Flutter SDK's `material/icons.dart` (base family).

### Changed
- The Flutter client resolves icon names through the generated `materialIcons` table instead of a hand-maintained ~20-icon switch, so any Material icon now renders.

## [0.5.0] - 2026-06-09

### Added
- `fugo upgrade` — self-update the CLI to the latest release via `go install …@latest` (pass a version to pin, e.g. `fugo upgrade v0.4.2`). On Windows the running binary is moved aside (`<exe>.old`) so `go install` can replace it.

### Changed
- Repo hygiene: consolidated `.gitignore` so scratch artifacts never land in the tree — throwaway `fugo init` demo projects, manual-test screenshots and captured run logs belong under `.scratch/` (root-level `*.png` / `*.out` / `*.err` are ignored too). Documented the repository layout — including why Go tests live beside their package rather than in a separate folder — in `AGENTS.md`.

## [0.4.2] - 2026-06-09

### Added
- `FloatingActionButton` now uses a unique hero tag per node, so an app can show multiple FABs (e.g. an increment and a decrement) without a Hero tag collision. Added the `remove` (minus) icon.

### Changed
- `fugo init`'s template is now a minimal, elegant counter: a centered count, an app bar titled "Fugo", and two FABs (decrement / increment). Nothing to tune — the shortest idiomatic Go + Flutter Material 3 demo.

## [0.4.1] - 2026-06-09

### Fixed
- A `FloatingActionButton` fired its `OnClick` repeatedly on its own. The client wrapped every app in an outer `Scaffold`, so an `fg.Scaffold` with a FAB became a **nested** Scaffold, which misroutes the FAB's gestures. The outer surface is now a plain `Material`, leaving the app's own `Scaffold` as the only one.

### Changed
- Cleaner, flatter look: the client flattens the Material 3 seed-tinted surfaces to a neutral background (white in light, near-black in dark); the seed still colors interactive elements (buttons, FAB, switches).
- `fugo init`'s counter template is now the canonical responsive Material app — a `Scaffold` with an app bar, a centered count, and a floating action button — so it fills the window instead of leaving a small centered cluster.

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
