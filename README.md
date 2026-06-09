<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/logo.svg">
  <img alt="Fugo" src="assets/logo.svg" width="64" height="64">
</picture>

# Fugo

**Server-Driven UI framework for desktop applications — write your logic in Go, render with Flutter.**

[![Go Version](https://img.shields.io/badge/Go-1.26.3-blue?logo=go)](https://go.dev)
[![Flutter](https://img.shields.io/badge/Flutter-3.24+-blue?logo=flutter)](https://flutter.dev)
[![gRPC](https://img.shields.io/badge/gRPC-bidirectional-purple)](https://grpc.io)
[![Protobuf](https://img.shields.io/badge/Protobuf-typed-orange)](https://protobuf.dev)
[![UDS](https://img.shields.io/badge/UDS-5%E2%80%9310%C2%B5s-brightgreen)](#)
[![License](https://img.shields.io/badge/license-MIT-green)](#)
[![Version](https://img.shields.io/badge/version-0.4.0-brightgreen)](VERSION)
[![go install](https://img.shields.io/badge/go%20install-cmd%2Ffugo-00ADD8?logo=go)](#installation)

---

## What is Fugo?

Fugo is a **local Server-Driven UI (SDUI)** framework that lets you build native desktop applications writing **exclusively in Go**. Business logic, state management, and routing live entirely in a Go process, while a precompiled Flutter engine acts as a pure rendering terminal — communicating over **Unix Domain Sockets** (TCP on Windows) via **gRPC** with **Protocol Buffers**.

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
fugo doctor      # checks Go, Flutter, protoc, gofumpt
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
`FilledButton`). Other native Material widgets: `fg.Card`, `fg.Scaffold` (with `.AppBar` / `.FAB`),
`fg.FloatingActionButton`, `fg.ListTile`, `fg.Chip`, and `fg.ProgressCircular` /
`fg.ProgressLinear`.

A bare `fg.Column` (or any intrinsically-sized root) auto-centers in the window; wrap a region in
`fg.Scaffold`/`fg.Container` to fill it instead. Tokens live under `Colors` (Primary, Surface,
OnSurface, Muted, Border, …), `Typography` (Heading/Body/Caption), `Spacing` (XS→XL), and
`Radius` (SM/MD/LG).

---

## Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| **Language** | Go 1.26+ | Goroutines, strong ecosystem, systems-level performance |
| **Rendering** | Flutter 3.24+ / Impeller | 60/120 fps native, world-class layout engine |
| **IPC Transport** | Unix Domain Sockets (TCP fallback on Windows) | 5-10µs latency, kernel-level throughput |
| **RPC** | gRPC bidirectional streaming | Typed contracts, health checking, keepalive |
| **Serialization** | Protocol Buffers (`google.golang.org/protobuf`) | Per-widget props marshaled as nested protobuf inside each node |
| **Wire updates** | Tree diff (ID/positional) | Only changed nodes stream as patches, never the full tree |
| **Process Mgmt** | `os/exec` + signals | Subprocess lifecycle, zombie prevention |
| **Window Mgmt** | `window_manager` | Cross-platform frameless windows, custom chrome |

---

## Current Status

**Version 0.4.0 — engine + widget API + transport + CLI + Flutter client are implemented and run end-to-end, the CLI is installable via `go install`, and the client renders native Material 3.**

- [x] Installable: `go install github.com/sazardev/fugo/cmd/fugo@latest` (generated protobuf bindings committed; builds on a clean fetch)
- [x] Native **Material 3** (light by default), seeded from `fg.Theme`; Material button variants (Filled/Tonal/Outlined/Text/Elevated/Icon) + Card/Scaffold/FAB/ListTile/Chip/Progress
- [x] Diffing engine, reconciler, 60 fps scheduler with priority (`Update` / `UpdateNow`)
- [x] gRPC transport (UDS / TCP on Windows), health check, keepalive, opt-in auth token
- [x] 36+ widgets in `fg/` with a fluent, prefix-free API + a `Theme` system
- [x] Flutter render client (background gRPC isolate, widget registry, auto-reconnect)
- [x] CLI: `fugo init` (templates) / `run` (`--watch`) / `build` / `doctor` / `widgets` / `upgrade` (self-update)
- [x] Runtime window control (`Context.Window()`), `window_manager`-backed
- [x] OS host services: clipboard (`Context.Clipboard()`), native file dialogs (`Context.Files()`)
- [x] Performance: object-pooled diff, GC tuning (`FUGO_GOGC` / `FUGO_GOMEMLIMIT`), Go + Dart benchmarks with a CI perf gate

See [ROADMAP](./ROADMAP/) and [SPEC.md](./SPEC.md) for the full design vision. **Note:** the
roadmap describes a FlatBuffers transport; the shipped implementation uses standard
**Protocol Buffers** (`google.golang.org/protobuf`) instead — per-widget props are a protobuf
message marshaled into each node's `bytes` field. `CLAUDE.md` is the canonical, up-to-date guide.

---

## Packages

```
fugo/                   # App, Context, lifecycle (RunStandalone, scheduler)
├── fg/                 # Declarative widgets (fg.Container, fg.Text, fg.Button, ...) + Theme
├── style/              # Styling primitives (Color, EdgeInsets, TextStyle, Border, ...)
├── engine/             # Diffing engine, Reconciler, Scheduler (16ms tick)
├── transport/          # gRPC server (UDS/TCP), health, keepalive
├── supervisor/         # Flutter subprocess lifecycle, signals
└── flutter_client/     # Precompiled Flutter rendering client
```

---

## CLI

```bash
fugo init <name>          # Scaffold a project (use --template app for a themed multi-page starter)
fugo run                  # Build + run; auto-builds the Flutter client the first time if it's missing
fugo run --watch          # Hot reload: rebuild the Go server on .go changes; the window stays open
fugo build                # Build + bundle the Flutter client into a self-contained dist/
fugo doctor               # Verify the dev environment (Go, Flutter, protoc, gofumpt)
fugo upgrade              # Self-update the CLI to the latest release (go install ...@latest)
fugo --version            # Print version information
```

**Hot reload** keeps the Flutter window open and reconnects after each Go rebuild (the in-memory
state still resets — full state restore would need a managed-state layer and is not implemented yet).

**Stateful components** are an alternative to a buildUI closure — implement `Render(ctx)` and pass
the value to `fugo.RunComponent`. **Routing** supports `:params` (e.g. `/user/:id`), read with
`ctx.Param("id")`. Set **`FUGO_AUTH=1`** to mint a per-run token that hardens the local transport.

**OS host services** run on the client and answer asynchronously: `ctx.Clipboard().Write/Read`,
`ctx.Files().Open/Save(fg.FileDialog{...}, func(path string){...})`. The callback runs on the
event goroutine, so mutate widgets and call `ctx.Update()` from it like any handler. For frameless
windows, wrap a region in **`fg.WindowDragArea(...)`** to make it drag the window, and use
**`fg.AnimatedPositioned(...)`** inside a `Stack` to animate a child between positions.

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
