<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/logo.svg">
  <img alt="Fugo" src="assets/logo.svg" width="64" height="64">
</picture>

# Fugo

**Server-Driven UI framework for desktop applications — write your logic in Go, render with Flutter.**

[![Go Version](https://img.shields.io/badge/Go-1.26.3-blue?logo=go)](https://go.dev)
[![Flutter](https://img.shields.io/badge/Flutter-3.24+-blue?logo=flutter)](https://flutter.dev)
[![gRPC](https://img.shields.io/badge/gRPC-bidirectional-purple)](https://grpc.io)
[![FlatBuffers](https://img.shields.io/badge/FlatBuffers-zero--copy-orange)](https://flatbuffers.dev)
[![UDS](https://img.shields.io/badge/UDS-5%E2%80%9310%C2%B5s-brightgreen)](#)
[![License](https://img.shields.io/badge/license-MIT-green)](#)
[![Version](https://img.shields.io/badge/version-0.1.0--alpha-yellow)](VERSION)

---

## What is Fugo?

Fugo is a **local Server-Driven UI (SDUI)** framework that lets you build native desktop applications writing **exclusively in Go**. Business logic, state management, and routing live entirely in a Go process, while a precompiled Flutter engine acts as a pure rendering terminal — communicating over **Unix Domain Sockets** via **gRPC** + **FlatBuffers**.

```
┌──────────────────────┐     IPC (UDS)       ┌──────────────────────┐
│      Go Process      │◄══════════════════►│   Flutter Process     │
│                      │  gRPC + FlatBuffers │                      │
│  ┌────────────────┐  │                      │  ┌────────────────┐  │
│  │ Business Logic   │  │   Widget Tree Diff  │  │ Widget Registry│  │
│  │ Virtual DOM     │──┼────────────────────►│  │ Render Pipeline│  │
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
| JSON parsing kills frame budgets | FlatBuffers **zero-copy** serialization |

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/sazardev/fugo"
    "github.com/sazardev/fugo/ui"
    "github.com/sazardev/fugo/style"
)

func main() {
    app := fugo.NewApp(fugo.AppOptions{
        Title:  "Fugo Desktop",
        Width:  800,
        Height: 600,
    })

    darkTheme := style.New(
        style.BgColor("#121212"),
        style.TextColor("#FFFFFF"),
    )

    app.Run(func(ctx *fugo.Context) ui.Widget {
        counter := 0

        counterText := ui.Text("0").
            FontSize(48).
            Style(darkTheme)

        incrementBtn := ui.Button("Increment").
            OnClick(func(e ui.Event) {
                counter++
                counterText.SetText(fmt.Sprint(counter))
                ctx.Update()
            }).
            Padding(16).
            BorderRadius(4)

        return ui.Container(
            ui.Center(
                ui.Column(counterText, incrementBtn).WithGap(24),
            ),
        ).Style(darkTheme).Fill()
    })
}
```

---

## Tech Stack

| Layer | Technology | Why |
|-------|-----------|-----|
| **Language** | Go 1.26+ | Goroutines, strong ecosystem, systems-level performance |
| **Rendering** | Flutter 3.24+ / Impeller | 60/120 fps native, world-class layout engine |
| **IPC Transport** | Unix Domain Sockets (TCP fallback on Windows) | 5-10µs latency, kernel-level throughput |
| **RPC** | gRPC bidirectional streaming | Typed contracts, health checking, keepalive |
| **Serialization** | FlatBuffers | Zero-copy reads in Dart, no heap allocations per frame |
| **gRPC Codec** | vtprotobuf | Zero-alloc marshal/unmarshal in Go |
| **Process Mgmt** | `os/exec` + signals | Subprocess lifecycle, zombie prevention |
| **Window Mgmt** | `window_manager` | Cross-platform frameless windows, custom chrome |

---

## Current Status

**Version 0.1.0 — Infrastructure phase complete.**

- [x] CI pipeline (lint, vet, build, test, format)
- [x] Git hooks via Lefthook (pre-commit, pre-push)
- [x] Go module skeleton (`github.com/sazardev/fugo`)
- [x] `.golangci.yml` with 80+ linters

Upcoming: Transport layer → Core SDK → Flutter client → Public API.

See [ROADMAP](./ROADMAP/) and [SPEC.md](./SPEC.md) for the full plan.

---

## Packages

```
fugo/                   # App, Context, lifecycle
├── ui/                 # Declarative widgets (Container, Text, Button, ...)
├── style/              # Styling primitives (Font, Color, Border, ...)
├── engine/             # Virtual DOM, Diffing engine, Reconciler, Scheduler
├── transport/          # gRPC server, FlatBuffers codec
├── supervisor/         # Flutter subprocess lifecycle, signals
└── flutter_client/     # Precompiled Flutter rendering client
```

---

## CLI

```bash
fugo init <name>     # Scaffold a new project
fugo run             # Start Go server + Flutter engine
fugo build           # Package into a single executable
fugo doctor          # Verify development environment
fugo version         # Print version information
```

---

## Design Principles

- **Go is the source of truth** — all logic, state, and routing in Go
- **No shared memory** — strict message passing via gRPC
- **Zero-copy wherever possible** — FlatBuffers for the widget tree
- **Opinionated on state, unopinionated on visuals** — one state model, any design system
- **Performance is a requirement, not an afterthought**
- **Terminal-native DX** — `fugo init` → `fugo run` → `fugo build`

---

## License

MIT — see [LICENSE](./LICENSE).
