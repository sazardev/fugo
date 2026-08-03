<img alt="Fugo" src="assets/logo.svg" width="64" height="64">

# Fugo

**Server-Driven UI framework for desktop applications — write your logic in Go, render with Flutter.**

[![Go Version](https://img.shields.io/badge/Go-1.26.3-blue?logo=go)](https://go.dev)
[![Flutter](https://img.shields.io/badge/Flutter-3.44.8-blue?logo=flutter)](FLUTTER_VERSION)
[![gRPC](https://img.shields.io/badge/gRPC-bidirectional-purple)](https://grpc.io)
[![Protobuf](https://img.shields.io/badge/Protobuf-typed-orange)](https://protobuf.dev)
[![UDS](https://img.shields.io/badge/UDS-5%E2%80%9310%C2%B5s-brightgreen)](#)
[![License](https://img.shields.io/badge/license-MIT-green)](#)
[![Version](https://img.shields.io/badge/version-3.44.8--fugo.0-brightgreen)](VERSION)
[![go install](https://img.shields.io/badge/go%20install-cmd%2Ffugo-00ADD8?logo=go)](#installation)

**[🌐 Website](https://sazardev.github.io/fugo/)** · **[📖 The complete guide](docs/FUGO_IDIOMATICO.md)** · **[📝 Changelog](CHANGELOG.md)**

> Fugo's version tracks the exact Flutter release it targets — `X.Y.Z` always matches the Flutter SDK the precompiled client is built against, with a `-fugo.N` suffix for fugo-only fixes against that same Flutter version. See `FLUTTER_VERSION` / `CLAUDE.md` for details.

---

## What is Fugo?

Fugo is a **local Server-Driven UI (SDUI)** framework that lets you build native desktop applications writing **exclusively in Go**. Business logic, state management, and routing live entirely in a Go process, while a precompiled Flutter engine acts as a pure rendering terminal — communicating over **Unix Domain Sockets** (TCP on Windows) via **gRPC** with **Protocol Buffers**.

> 📖 **[docs/FUGO_IDIOMATICO.md](docs/FUGO_IDIOMATICO.md)** is the complete, code-verified guide (in Spanish) to the whole framework: architecture, the full 83-widget catalog, theming, idiomatic state management (closures vs. the opt-in `fg.Store[S]`), host services, platform support, and packaging. It's the canonical reference — this README stays a pitch and quick start.

```
┌──────────────────────┐     IPC (UDS/TCP)   ┌──────────────────────┐
│      Go Process      │◄══════════════════►│   Flutter Process     │
│                      │   gRPC + Protobuf   │                      │
│  ┌────────────────┐  │                      │  ┌────────────────┐  │
│  │ Business Logic   │  │   Widget Tree Diff  │  │ Widget Registry│  │
│  │ Retained Tree   │──┼────────────────────►│  │ Render Pipeline│  │
│  │ Diffing Engine  │  │                     │  │ Event Debouncer│  │
│  │ gRPC Server     │  │   User Events       │  │ gRPC Client    │  │
│  └────────────────┘  │◄─────────────────────│  └────────────────┘  │
└──────────────────────┘                      └──────────────────────┘
```

**Go is the absolute source of truth.** Flutter is a dumb terminal — no business logic, no state, just pixels at 60/120 fps via Impeller.

---

## Why Fugo?

| Problem | Fugo's Answer |
|---------|---------------|
| Electron apps consume >150MB RAM | Native rendering via Flutter/Impeller, no Chromium |
| Go GUI libraries (Fyne, Gio) lack widget ecosystem | Flutter's world-class typography, layout, animations |
| Flutter forces you into Dart for everything | Write all logic in Go, use any Go library |
| Remote SDUI suffers 50-200ms network latency | Local IPC via UDS: **5-10µs** round-trip |
| JSON parsing kills frame budgets | Compact **Protobuf** framing; only diffs cross the wire |

---

## Installation

Install the `fugo` CLI straight from source (requires **Go 1.26+**):

```bash
go install github.com/sazardev/fugo/cmd/fugo@latest
```

This drops the `fugo` binary in `$(go env GOPATH)/bin` — make sure that's on your `PATH`, then:

```bash
fugo --version
fugo doctor      # checks the toolchain + (in a project) its health
```

The generated protobuf bindings are **committed**, so a clean module fetch compiles without `protoc` or any code-gen step.

> **Rendering prerequisite.** `fugo init`, `fugo doctor` and `fugo widgets` work standalone. But because Fugo renders through a precompiled **Flutter** client, `fugo run` / `fugo build` additionally require the [Flutter SDK](https://docs.flutter.dev/get-started/install) with desktop support enabled. The CLI builds the render client on first `run`; alternatively, point **`FUGO_FLUTTER_BINARY`** at a prebuilt client binary. `go install` ships the Go CLI only — not the Flutter engine.

**From a clone** (e.g. to hack on the framework):

```bash
git clone https://github.com/sazardev/fugo && cd fugo
go build ./cmd/fugo      # or: make cli   (Go bindings are committed; no protoc needed)
```

---

## Quick Start

```go
package main

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

func main() {
	fugo.RunStandalone(fugo.AppOptions{
		Title:  "Fugo Desktop",
		Width:  800,
		Height: 600,
	}, buildUI)
}

func buildUI(ctx *fugo.Context) fg.Widget {
	counter := 0
	counterText := fg.Text("0").FontSize(48)

	incBtn := fg.Button("+").
		BgColor(fg.Hex("#10B981")).
		FontSize(20).
		OnClick(func(_ fg.Event) {
			counter++
			counterText.SetText(strconv.Itoa(counter))
			ctx.Update() // mark dirty → diff → patch streamed to Flutter
		})

	return fg.Container(
		fg.Column(
			counterText,
			fg.SizedBox(0, 16),
			incBtn,
		),
	).BgColor(fg.Hex("#1A1A2E")).Pad(fg.EdgeAll(24))
}
```

The widget tree is **built once and retained**. Event handlers are Go closures that mutate
widget fields in place (e.g. `counterText.SetText(...)`) and call `ctx.Update()`; the scheduler
re-walks the same tree each frame, diffs it, and streams only the patches.

> Constructors are **prefix-free**: `fg.Text(...)`, `fg.Button(...)`, `fg.Container(...)` —
> not `NewText`. Each returns a concrete `*fg.TextWidget` / `*fg.ButtonWidget` / … with
> chainable setters.

---

## Theming & Material 3

Fugo renders with **Material 3** and a **light** color scheme by default. The active `fg.Theme`'s
primary color seeds Flutter's `ColorScheme.fromSeed`, so widgets get native M3 colors
automatically — a `fg.FilledButton` looks like a real filled button without setting any color.
Per-widget setters still override the theme.

```go
fg.UseTheme(fg.DarkTheme()) // light is active by default — call before RunStandalone

t := fg.CurrentTheme()
fg.Text("Title").FontSize(t.Typography.Heading)
fg.SizedBox(0, t.Spacing.LG)
```

**Buttons** mirror Material 3 — `fg.FilledButton`, `fg.FilledTonalButton`, `fg.OutlinedButton`,
`fg.TextButton`, `fg.ElevatedButton`, `fg.IconButton` (and `fg.Button`, an alias of
`FilledButton`). Other native Material widgets: `fg.Card`, `fg.Scaffold`, `fg.AppBar`
(title + `.Leading` / `.Actions`), `fg.FloatingActionButton`, `fg.ListTile`, `fg.Chip`, and
`fg.ProgressCircular` / `fg.ProgressLinear`, `fg.NavigationBar`, and `fg.Tabs` (a `TabBar` +
`TabBarView`, switched client-side) — plus `fg.Tooltip`, `fg.Badge`, `fg.CircleAvatar`,
`fg.SegmentedButton`, `fg.Spacer`, `fg.AspectRatio`, `fg.ClipRRect`, `fg.FittedBox`, `fg.Flexible`,
`fg.ExpansionTile`, `fg.PopupMenuButton`, `fg.RichText`, `fg.DataTable`, and `fg.Stepper` — plus
forms (`fg.Form`), drag-and-drop (`fg.Draggable`/`fg.DragTarget`), swipe-to-dismiss
(`fg.Dismissible`), collapsing headers (`fg.SliverScaffold`), and more; see
[docs/FUGO_IDIOMATICO.md](docs/FUGO_IDIOMATICO.md) for the complete catalog. A scaffold composes
the essentials — an app bar, the body, a FAB, a slide-in `.Drawer`, and a bottom `.BottomBar`:

```go
fg.Scaffold(body).
    AppBar(fg.AppBar("Inbox").Actions(fg.IconButton(fg.Icons.Search))).
    Drawer(fg.Column(fg.ListTile("Home").Leading(fg.Icons.Home))).
    BottomBar(fg.NavigationBar().
        Item(fg.Icons.Home, "Home").
        Item(fg.Icons.Person, "Profile").
        OnChange(func(e fg.Event) { /* e.Data is the selected index */ })).
    FAB(fg.FloatingActionButton(fg.Icons.Add))
```

A bare `fg.Column` (or any intrinsically-sized root) auto-centers in the window; wrap a region in
`fg.Scaffold`/`fg.Container` to fill it instead. Tokens live under `Colors` (Primary, Surface,
OnSurface, Muted, Border, …), `Typography` (Heading/Body/Caption), `Spacing` (XS→XL), and
`Radius` (SM/MD/LG).

### Skip the boilerplate: `fg.Icons`, `fg.Colors`, `fg.TextSize`

Use Flutter's constants instead of hand-written strings, hex, and magic numbers:

```go
fg.IconButton(fg.Icons.Favorite)                       // ~2,200 Material icons: fg.Icons.Home, .Coffee, .Settings…
fg.Container(child).BgColor(fg.Colors.Amber)           // the Material palette: fg.Colors.Blue, .RedAccent, .Grey800…
fg.Text("Title").FontSize(fg.TextSize.HeadlineMedium)  // the M3 type scale: .DisplayLarge, .BodyMedium…
```

`fg.Icons.*` mirrors Flutter's `Icons` (generated from the installed SDK via `go run ./cmd/gen-icons`); `fg.Colors.*` mirrors `Colors`; `fg.TextSize.*` is the Material 3 type scale.

---

## Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| **Language** | Go 1.26+ | Goroutines, strong ecosystem, systems-level performance |
| **Rendering** | Flutter (pinned, see `FLUTTER_VERSION`) / Impeller | 60/120 fps native, world-class layout engine |
| **IPC Transport** | Unix Domain Sockets (TCP fallback on Windows) | 5-10µs latency, kernel-level throughput |
| **RPC** | gRPC bidirectional streaming | Typed contracts, health checking, keepalive |
| **Serialization** | Protocol Buffers (`google.golang.org/protobuf`) | Per-widget props marshaled as nested protobuf inside each node |
| **Wire updates** | Tree diff (ID/positional) | Only changed nodes stream as patches, never the full tree |
| **Process Mgmt** | `os/exec` + signals | Subprocess lifecycle, zombie prevention |
| **Window Mgmt** | `window_manager` | Cross-platform frameless windows, custom chrome |

### Platform support

| Platform | Status |
|---|---|
| Linux (X11) | ✅ Supported |
| Linux (Wayland — Hyprland, Sway, GNOME, etc.) | ✅ Works via XWayland (verified on Hyprland); Flutter's Linux embedder isn't natively Wayland yet |
| Windows | ✅ Supported |
| macOS | Untested, should build (UDS transport, same as Linux) |
| Android / iOS | ❌ Out of scope — the subprocess + Unix-socket architecture doesn't fit a mobile sandbox |

---

## Current Status

**The engine, widget API, transport, CLI, and Flutter client are implemented and run end-to-end** — installable via `go install`, rendering native Material 3, with 83 widgets in `fg/`. Highlights:

- [x] Installable: `go install github.com/sazardev/fugo/cmd/fugo@latest` (generated protobuf bindings committed; builds on a clean fetch)
- [x] Native **Material 3** (light/dark/`Theme.FollowSystem`), seeded from `fg.Theme`; the full button family + Card/Scaffold/AppBar/FAB/ListTile/Chip/Progress/DataTable/Form/SliverScaffold and more
- [x] Diffing engine, reconciler, 60 fps scheduler with priority (`Update` / `UpdateNow`)
- [x] gRPC transport (UDS / TCP on Windows), health check, keepalive, opt-in auth token
- [x] **83 widgets** in `fg/` with a fluent, prefix-free API, a `Theme` system, and an opt-in `fg.Store[S]` for shared state
- [x] Flutter render client (background gRPC isolate, widget registry, auto-reconnect)
- [x] CLI: `fugo init` (templates) / `run` (hot reload by default) / `build` (+ Linux packaging) / `doctor` (`--fix`) / `autostart` / `widgets` / `upgrade`
- [x] Runtime window control (`Context.Window()`), keyboard shortcuts, focus requests, file drag-and-drop
- [x] OS host services: clipboard, native file dialogs, native notifications
- [x] Imperative overlays: `ctx.ShowSnackBar(...)`, `ctx.ShowDialog(...)`, date/time pickers
- [x] Performance: object-pooled diff, GC tuning (`FUGO_GOGC` / `FUGO_GOMEMLIMIT`), Go + Dart benchmarks with a CI perf gate

**[docs/FUGO_IDIOMATICO.md](docs/FUGO_IDIOMATICO.md) is the up-to-date, exhaustive picture** — this
list is a snapshot, that guide isn't. See [ROADMAP](./ROADMAP/) and [SPEC.md](./SPEC.md) for the
original design vision. **Note:** the roadmap describes a FlatBuffers transport; the shipped
implementation uses standard **Protocol Buffers** (`google.golang.org/protobuf`) instead —
per-widget props are a protobuf message marshaled into each node's `bytes` field. `CLAUDE.md` is
the canonical, up-to-date engineering guide.

---

## Packages

```
fugo/                   # App, Context, lifecycle (RunStandalone, scheduler)
├── fg/                 # Declarative widgets (fg.Container, fg.Text, fg.Button, ...) + Theme
├── style/              # Styling primitives (Color, EdgeInsets, TextStyle, Border, ...)
├── engine/             # Diffing engine, Reconciler, Scheduler (16ms tick)
├── transport/          # gRPC server (UDS/TCP), health, keepalive
├── supervisor/         # Flutter subprocess lifecycle, signals
├── cmd/fugo/           # The fugo CLI (init/run/build/doctor/autostart/widgets/...)
├── cmd/fugo-lsp/       # A from-scratch Language Server for fg — hover/definition/completion
├── fugovet/            # Fugo's opinionated static analyzer (go vet-compatible)
└── flutter_client/     # Precompiled Flutter rendering client
```

---

## CLI

```bash
fugo init <name>          # Scaffold a project (use --template app for a themed multi-page starter)
fugo run                  # Build + run; hot-reloads on .go changes (window stays open). Auto-builds the Flutter client the first time.
fugo run --no-watch       # Build and run once, without hot reload
fugo build                # Build + bundle the Flutter client into a self-contained dist/ (Linux: also writes a
                          #   .desktop entry, install.sh, and an AUR PKGBUILD template alongside it)
fugo doctor               # Check the toolchain; inside a project, validate fugo.toml + structure + that it compiles
fugo autostart enable     # Launch this app at login (XDG autostart on Linux, Run key on Windows)
fugo upgrade              # Self-update the CLI to the latest release (go install ...@latest)
fugo --version            # Print version information
```

`fugo init` scaffolds a **recommended layout** and initializes a git repo (initial commit; skip with `--no-git`):

```
myapp/
├─ main.go        # entrypoint: sets the theme, then fugo.RunStandalone(fugo.ConfigOptions("fugo.toml"), ui.Build)
├─ ui/            # your screens (package ui); ui.Build is the root widget
│  └─ home.go
├─ fugo.toml      # window title/size + gRPC address — read by the CLI and the app
├─ bin/           # dev builds        (gitignored)
├─ dist/          # release bundle    (gitignored)
├─ logs/          # fugo run → logs/run.log (gitignored)
├─ README.md
└─ .gitignore
```

Edit **`fugo.toml`** to change the window or server address — no recompiling the config into Go:

```toml
name = "myapp"

[window]
title  = "My App"
width  = 800
height = 600

[server]
addr = "127.0.0.1:9510"   # fugo run uses this unless you pass --addr
```

**Hot reload is on by default**: `fugo run` watches `.go` files and rebuilds the Go server on every
change while the Flutter window stays open and reconnects (~500ms), so your edits show up live. The
in-memory state resets across reloads — full state restore would need a managed-state layer and is
not implemented yet. Use `fugo run --no-watch` for a single build-and-run.

**Stateful components** are an alternative to a buildUI closure — implement `Render(ctx)` and pass
the value to `fugo.RunComponent`. **Routing** supports `:params` (e.g. `/user/:id`), read with
`ctx.Param("id")`. Set **`FUGO_AUTH=1`** to mint a per-run token that hardens the local transport.

**OS host services** run on the client and answer asynchronously: `ctx.Clipboard().Write/Read`,
`ctx.Files().Open/Save(fg.FileDialog{...}, func(path string){...})`, `ctx.Notifications().Show(title, body)`,
`ctx.RequestFocus(field)`, `ctx.RegisterShortcuts(map[string]func(){...})`, and `ctx.OnFileDrop(func(paths []string){...})`.
Callbacks run on the event goroutine, so mutate widgets and call `ctx.Update()` from them like any
handler. For frameless windows, wrap a region in **`fg.WindowDragArea(...)`** to make it drag the
window, and use **`fg.AnimatedPositioned(...)`** inside a `Stack` to animate a child between positions.
For state shared across more than a couple of widgets, `fg.Store[S]` is an opt-in generic store
(`Get`/`Update(fn)`/`Subscribe`) that composes with `ctx.Update()` rather than replacing it.

**Overlays** are imperative, driven from Go over the same command channel: `ctx.ShowSnackBar("Saved")`
(snackbar), `ctx.ShowDialog("Title", "Message")` (alert dialog), and `ctx.ShowBottomSheet("Title",
"Message")` (bottom sheet). The native pickers return a value to a callback: `ctx.PickDate(func(d string){…})`
(ISO `YYYY-MM-DD`) and `ctx.PickTime(func(t string){…})` (`HH:MM`), empty if cancelled.

---

## Design Principles

- **Go is the source of truth** — all logic, state, and routing in Go
- **No shared memory** — strict message passing via gRPC
- **Stream only diffs** — ID/positional tree diffing, patches over gRPC, never full re-renders
- **Opinionated on state, themed by default, unopinionated on design system**
- **Performance is a requirement, not an afterthought**
- **Terminal-native DX** — `fugo init` → `fugo run` → `fugo build`

---

## License

MIT — see [LICENSE](./LICENSE).
