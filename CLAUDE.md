# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What Fugo is

A local **Server-Driven UI** framework for desktop. You write app logic, state, and routing **entirely in Go**; a precompiled Flutter binary acts as a dumb render terminal. The two processes talk over gRPC bidirectional streaming (TCP on Windows, UDS elsewhere). Go is the single source of truth — Flutter holds no business logic or state.

> **Note on stale docs:** `AGENTS.md` claims "zero application code exists yet" — this is **out of date**. The engine, widget API, transport, supervisor, CLI, and Flutter client are all implemented. The **ROADMAP** (notably `05_TRANSPORTE.md`) still describes FlatBuffers + vtprotobuf serialization, but the code uses **standard `google.golang.org/protobuf`** (see "Wire format" below). `README.md` and `SPEC.md` have been realigned to the code (Protobuf, `fg/` API); `CLAUDE.md` is canonical. Trust the code over the ROADMAP.

## Commands

Module: `github.com/sazardev/fugo`, Go 1.26.3. Formatter is **`gofumpt`** (not `gofmt`) — run `gofumpt -w .` before committing.

```sh
make test          # go test ./... -count=1 -race -shuffle=on -v
make lint          # golangci-lint run --timeout 10m ./...   (80+ linters via .golangci.yml + staticcheck)
make vet           # go vet ./...
make build         # build bin/fugo (CLI) with version/commit/date ldflags
make install       # install Lefthook git hooks (go tool lefthook install)

# Run a single test (use the package + -run regex):
go test ./engine/ -run TestDiff_Update -race -v
go test ./engine/ -run TestDiff -bench BenchmarkDiff -benchmem   # benchmarks live in *_bench_test.go
```

### Running the app end-to-end (Windows)

There is no single `make run` on Windows (`make run-spike` / `make flutter-build` are Linux-only). The flow is:

```sh
# 1. Build the Flutter render client once (output: flutter_client/build/windows/.../fugo_flutter_client.exe)
cd flutter_client && flutter build windows

# 2a. Run the bundled demo (router + every widget) — builds Go server + auto-spawns the Flutter binary:
make cli                 # builds bin/fugo.exe and bin/fugo-spike.exe
./bin/fugo-spike.exe

# 2b. Or scaffold and run a user app:
./bin/fugo.exe init myapp      # scaffolds main.go + ui/ package + fugo.toml + README + .gitignore + logs/,
                               # runs go mod init/tidy (+ replace to local fugo), and git init + initial commit
cd myapp && fugo run           # go build → launch app → spawn Flutter; add --watch to rebuild on .go change
```

**Generated project layout** (`fugo init`): a thin `main.go` (sets the theme, then `fugo.RunStandalone(fugo.ConfigOptions("fugo.toml"), ui.Build)`), a `ui` package holding the screens (`ui.Build` is the root), `fugo.toml` (`[window] title/width/height` + `[server] addr`, read by the CLI **and** the app via the dependency-free `config` package), `README.md`, `.gitignore`, and `logs/` (`fugo run` tees runtime logs to `logs/run.log`). Build output goes to `bin/` (dev) and `dist/` (release, `fugo build` — which also ships `fugo.toml` beside the binary). `fugo run` uses `[server] addr` as the default address unless `--addr` is passed.

`fugo doctor` checks for Go / Flutter / protoc / gofumpt. The Go process finds the Flutter binary via `FUGO_FLUTTER_BINARY`, then by searching up the tree for the fugo repo (see `findFlutterBinary` in `app.go` / `cmd/fugo-spike/main.go`). The gRPC address is `FUGO_ADDR` (default `127.0.0.1:9510`). The active `fg.Theme` is forwarded to the client as `FUGO_THEME_SEED` (primary color hex) + `FUGO_THEME_BRIGHTNESS` (`light`/`dark`), which seed the client's **Material 3** `ColorScheme.fromSeed` — **light by default** (call `fg.UseTheme(fg.DarkTheme())` before `RunStandalone` to change it).

### Regenerating protobuf code

`transport/proto/fugo/v1/fugo.proto` is the single source. `make proto` runs `protoc` to generate Go (`*.pb.go`, `*_grpc.pb.go`) **and** copies the proto into `flutter_client/proto/` to regenerate the Dart bindings under `flutter_client/lib/generated/`. After editing the proto you must run `make proto` and keep both sides in sync. Requires `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, and `protoc-gen-dart` on PATH.

> **`protoc-gen-dart` version pin (gotcha).** The Dart `protobuf` runtime is pinned to `^3.1.0` in `flutter_client/pubspec.yaml`, so the generator must be **`protoc_plugin` 21.1.x** (`dart pub global activate protoc_plugin 21.1.2`). A newer plugin (e.g. 25.x) emits code for the `protobuf` 6.x runtime — calls like `$_clearField` / `$_initByValueList` and `PbList` return types — which will not compile against 3.1.0. If `flutter analyze` floods with "method `$_clearField` isn't defined", you regenerated with the wrong plugin version.

The Go bindings (`*.pb.go`) are **committed to the repo** so that `go install github.com/sazardev/fugo/cmd/fugo@latest` and any clean-checkout `go build` work without `protoc` (the module proxy / a fresh clone never runs `make proto`). `make proto` keeps them **gofumpt-clean** (it runs `gofumpt -w` on the generated Go), so the `format` CI job — which checks the committed tree without regenerating — passes. The **Dart** bindings (`*.pb.dart` under `flutter_client/lib/generated/`) and the copied `flutter_client/proto/` stay **gitignored**; they're produced by `flutter build`. CI still regenerates the Go bindings via the `./.github/actions/gen-proto` composite action (lint/vet/build/test/bench jobs) to catch proto drift; golangci-lint excludes generated files by header, so linting is unaffected.

## Architecture & data flow

The render loop spans the Go process (`app.go` + packages below) and the Flutter client (`flutter_client/lib/`).

```
buildUI(ctx) ──> fg.Widget tree (built ONCE, retained)
                      │
   event handlers mutate widget structs in place, then call ctx.Update()
                      │  marks scheduler dirty
   engine.Scheduler (16ms tick / 60fps) ── if dirty ──> App.flush()
                      │
   fg.BuildTreeWithMerge: re-walk retained tree -> WidgetTree (proto)
                      │
   engine.Diff(oldTree, newTree) -> []Patch   (CREATE/UPDATE/DELETE/REPLACE/REORDER)
                      │
   engine.Reconciler.SendPatches -> gRPC RenderStream -> Flutter
                      │
   Flutter applies patches to its flat WidgetNode map, rebuilds via WidgetRegistry
                      │
   user interaction -> ClientEvent (debounced 16ms, except onClick/onTap/onSubmit/onLongPress) -> Go App.HandleEvent
```

**Key mental model:** the widget tree is built **once** in `App.Run` and **retained**. Event handlers are Go closures that mutate widget struct fields directly (e.g. `counterText.SetText("5")`) and then call `ctx.Update()`. The scheduler re-walks that *same* retained tree every frame when dirty, re-marshals props, diffs against the previous proto snapshot, and streams only the patches. This is **not** a React-style "rebuild from scratch each render."

### Packages

| Package | Role |
|---|---|
| `fugo` (root, `app.go`) | `App`, `Context`, lifecycle. `Run` builds the tree + starts scheduler; `RunStandalone` also starts the gRPC server, tunes the GC (`runtime_tuning.go`, `FUGO_GOGC`/`FUGO_GOMEMLIMIT`), and spawns Flutter. Owns the `handlers` map (nodeID → Widget) and routes `ClientEvent`s. `Context` exposes `Update`/`UpdateNow` (priority), `Window()` (runtime window control), and the OS host services in `host.go`: `Clipboard()` and `Files()` (native file dialogs) — async requests correlated by id and answered with a `"host"` ClientEvent. |
| `fg/` | The declarative widget API — **prefix-free** constructors (`fg.Text`, `fg.Container`, `fg.Button`, `fg.Router`, ...), **not** `NewText`. Each returns a concrete `*fg.TextWidget` / `*fg.ButtonWidget` / … with chainable setters. **This is the active (and only) widget package** — imported by `app.go` and both `cmd/`s. It also re-exports the `style`/theme helpers (`fg.Hex`, `fg.EdgeAll`, `fg.DarkTheme`, `fg.CurrentTheme`, ...). Each widget implements `Widget`; `walkNodes` assigns IDs depth-first and marshals its `*Props` proto into `WidgetNode.Props`. |
| `style/` | Styling primitives: `Color` (`style.Hex(...)`), `TextStyle`, `EdgeInsets` (`style.EdgeAll`), `Border`, font weights. (`fg` re-exports the common ones.) |
| `config/` | Dependency-free (stdlib-only) loader for `fugo.toml` (a small fixed TOML subset: `name`, `[window]`, `[server]`). Shared by the runtime (`fugo.ConfigOptions` in `config_options.go`) and the CLI; kept free of the gRPC stack so the CLI stays light. |
| `engine/` | `Diff` (ID/positional diff → patches; pools the per-frame lookup map via `sync.Pool`, with a zero-alloc no-change fast path), `Reconciler` (wraps the gRPC stream, buffers payloads until a client connects; `SendWindowCommand`/`SendHostCommand` for out-of-band ops), `Scheduler` (dirty-flag + ticker coalescing updates to one flush per frame; `EnqueueNow` wakes the loop immediately for latency-sensitive updates). |
| `transport/` | gRPC server (`StartServer`), health check, keepalive. UDS when addr has no `:`, TCP otherwise; adapts the stream into `engine.RenderStream`. |
| `supervisor/` | Spawns/monitors the Flutter subprocess (`StartFlutter`), forwards `FUGO_ADDR`, handles graceful shutdown. |
| `cmd/fugo/` | The `fugo` CLI: `init`, `run` (+`--watch`), `build`, `doctor`, `widgets`, `upgrade` (self-update via `go install …@latest`). |
| `cmd/fugo-spike/` | Demo/integration harness — a hand-written app exercising the router and every widget. Good reference for the widget API. |
| `flutter_client/` | Dart render client. `grpc_isolate.dart` runs the gRPC client on a separate isolate (auto-reconnect w/ backoff) and ships raw proto bytes via `SendPort`. `fugo_renderer.dart` keeps a flat `Map<int, WidgetNode>` and applies patches. `registry.dart` maps each `WidgetType` to a Flutter widget by decoding the embedded props. `events.dart` debounces outbound events. |

### Wire format

`WidgetNode.props` is a `bytes` field containing a **protobuf-marshaled** per-widget props message (`TextProps`, `ButtonProps`, ...) — i.e. protobuf nested inside protobuf, marshaled with `proto.Marshal` on the Go side and decoded with `*.fromBuffer(node.props)` on the Dart side. The whole `RenderPayload` is a normal gRPC protobuf message. There is no FlatBuffers/vtprotobuf in the codebase despite README claims.

When adding a widget you must touch **four places**: the proto (`WidgetType` enum + a `*Props` message), regenerate (`make proto`), add the Go widget in `fg/` (implement `walkNodes`, marshal props), and add the Dart builder in `flutter_client/lib/registry.dart`.

**Constants (Flutter-style).** To avoid hand-written strings/hex/sizes, `fg` ships namespaced banks: **`fg.Icons.Home`** (the full Material icon set, ~2,200 base icons), **`fg.Colors.Amber`** (the Material palette), and **`fg.TextSize.HeadlineMedium`** (the M3 type scale). `fg.Icons.X` is just the icon's Flutter name (a string), which the client resolves to `IconData` via the generated `flutter_client/lib/icons_gen.dart` map; both that map and `fg/icons_gen.go` are produced by **`go run ./cmd/gen-icons`** (parses the installed Flutter SDK's `material/icons.dart`, base family only — re-run it after a Flutter SDK upgrade). `fg.Colors`/`fg.TextSize` are hand-maintained in `fg/colors.go` / `fg/textsize.go`.

**Material 3 / vanilla styling.** The client is M3-native, so widgets do **not** inject opinionated hex colors by default — `Text`, `Container`, and `Button` only marshal a color when the user calls a setter (tracked by a `*Set bool` flag; an unset color is sent as `""` so proto3 omits it and the M3 `ColorScheme` styles it). Button variants are separate constructors (`FilledButton`/`FilledTonalButton`/`OutlinedButton`/`TextButton`/`ElevatedButton`/`IconButton`; `Button` = filled) carried by `ButtonProps.variant`. Core Material widgets: `Card`, `Scaffold` (`.AppBar`/`.FAB`), `FloatingActionButton`, `ListTile`, `Chip`, `ProgressCircular`/`ProgressLinear`. The renderer auto-centers an intrinsically-sized root (e.g. a bare `Column`, which defaults to `MainAxisSize.min`); roots whose type fills the viewport (`Scaffold`, `Container`, `ListView`, …) are not wrapped — see `_fillTypes` in `fugo_renderer.dart`.

## Conventions & workflow

- **Releases:** never hand-edit `VERSION` or `CHANGELOG.md`. Use `make release TYPE=patch|minor|major MSG="..."` (bumps VERSION, updates CHANGELOG, commits, tags). Pre-push hook fails if VERSION lacks a matching CHANGELOG entry.
- **PRs:** `make pr MSG="type: description"` (must be on `main`; creates branch, pushes, opens PR via `gh`). `make pr-merge PR=<n>` squash-merges and cleans up.
- **Git hooks (Lefthook):** pre-commit auto-fixes staged Go files (`golangci-lint --fast --fix`, vet, gofumpt check, mod tidy). Pre-push runs the full gate: golangci-lint, vet, staticcheck, build, `go test -race -shuffle=on`, mod tidy.
- **CI** (push/PR to `main`): five jobs — lint+staticcheck, vet, build, test (`-race -shuffle=on`), gofumpt format check.

## Repo quirks

- **`fg/` is the live (and only) widget package; there is no `ui/` directory.** Everything imports `github.com/sazardev/fugo/fg` (`app.go`, both `cmd/`s, the CLI scaffold templates, the README Quick Start). Constructors are **prefix-free** (`fg.Text(...)`, `fg.Button(...)`), **not** `New*`. The ROADMAP/SPEC docs (`06_CLI.md`, `07_API_GO.md`) reference a `ui` package and `ui.NewText`-style constructors — **those don't exist in the code and are aspirational**. Trust `fg/`.
- `ROADMAP/` (12 files) and `docs/` are the real design specs and are mostly **in Spanish**. `SPEC.md` is the detailed specification.
- Most `go.mod` dependencies are `// indirect` (pulled transitively by the Lefthook tool dependency and the CLI). Direct deps are gRPC and protobuf.
- `bin/*.exe` and `fugo*.exe` at the repo root are committed build artifacts.
