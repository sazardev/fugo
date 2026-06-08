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
[![Version](https://img.shields.io/badge/version-0.1.0--alpha-yellow)](VERSION)

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

## Theming

Fugo ships opinionated dark/light themes. The active theme feeds widget defaults (text color,
button color, radii, sizes); per-widget setters always override it.

```go
fg.UseTheme(fg.LightTheme()) // fg.DarkTheme() is active by default

t := fg.CurrentTheme()
fg.Text("Title").FontSize(t.Typography.Heading)
fg.Button("Save").BgColor(t.Colors.Success)
fg.SizedBox(0, t.Spacing.LG)
```

Tokens live under `Colors` (Primary, Surface, OnSurface, Muted, Border, …), `Typography`
(Heading/Body/Caption), `Spacing` (XS→XL), and `Radius` (SM/MD/LG).

---

## Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| **Language** | Go 1.26+ | Goroutines, strong ecosystem, systems-level performance |
| **Rendering** | Flutter 3.24+ / Impeller | 60/120 fps native, world-class layout engine |
| **IPC Transport** | Unix Domain Sockets (TCP fallback on Windows) | 5-10µs latency, kernel-level throughput |
| **RPC** | gRPC bidirectional streaming | Typed contracts, health checking, keepalive |
| **Serialization** | Protocol Buffers (`google.golang.org/protobuf`) | Per-widget props marshaled as nested protobuf inside each node |
| **Wire updates** | Keyed tree diff | Only changed nodes stream as patches, never the full tree |
| **Process Mgmt** | `os/exec` + signals | Subprocess lifecycle, zombie prevention |
| **Window Mgmt** | `window_manager` | Cross-platform frameless windows, custom chrome |

---

## Current Status

**Version 0.1.0 — engine + widget API + transport + CLI + Flutter client are implemented and run end-to-end.**

- [x] Diffing engine, reconciler, 60 fps scheduler
- [x] gRPC transport (UDS / TCP on Windows), health check, keepalive
- [x] 24 widgets in `fg/` with a fluent, prefix-free API + a `Theme` system
- [x] Flutter render client (background gRPC isolate, widget registry, auto-reconnect)
- [x] CLI: `fugo init` / `run` (`--watch`) / `build` / `doctor`

See [ROADMAP](./ROADMAP/) and [SPEC.md](./SPEC.md) for the full design vision. **Note:** the
roadmap/spec describe a FlatBuffers transport; the shipped implementation uses standard
**Protocol Buffers** instead.

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
fugo init <name>     # Scaffold a new project (main.go + go.mod)
fugo run             # Build Go server + spawn Flutter engine (add --watch to rebuild on change)
fugo build           # Build + bundle the Flutter client into a self-contained dist/
fugo doctor          # Verify development environment (Go, Flutter, protoc, gofumpt)
fugo --version       # Print version information
```

---

## Design Principles

- **Go is the source of truth** — all logic, state, and routing in Go
- **No shared memory** — strict message passing via gRPC
- **Stream only diffs** — keyed tree diffing, patches over gRPC, never full re-renders
- **Opinionated on state, themed by default, unopinionated on design system**
- **Performance is a requirement, not an afterthought**
- **Terminal-native DX** — `fugo init` → `fugo run` → `fugo build`

---

## License

MIT — see [LICENSE](./LICENSE).
