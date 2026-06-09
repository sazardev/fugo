# SPEC.md: Fugo Framework

> **Status note.** This is the original design specification. Two things diverge
> from the shipped code, on purpose: the transport uses standard **Protocol
> Buffers** (not FlatBuffers), and the widget package is **`fg/`** with
> prefix-free constructors (`fg.Text(...)`, not `ui.NewText(...)`). The code
> example in §5 below reflects the real API. `CLAUDE.md` is the canonical,
> up-to-date guide; when this document and the code disagree, the code wins.

## 1. Overview

**Fugo** is a high-performance Server-Driven UI (SDUI) framework for desktop application development. Its primary purpose is to allow developers to build complex and fluid user interfaces by writing **exclusively in Go**, delegating graphical rendering to a precompiled Flutter engine that remains invisible to the end user.

The framework prioritizes a seamless Developer Experience (DX) from the terminal, maximum execution speed via binary protocols, and a structurally agnostic design.

---

## 2. Project Scope

### What Fugo WILL Do

- **Decoupled Execution:** It will keep business logic and application state strictly within the Go process.
- **Native Rendering:** It will use Flutter (via Impeller/Canvas) in a separate subprocess to draw the interface at 60/120fps.
- **Real-Time Synchronization:** It will establish a bidirectional channel to transmit the UI tree (from Go) and user events (from Flutter) with minimal latency.
- **Agnostic Styling:** It will provide atomic design primitives (colors, typography, borders) without imposing predefined design systems.
- **Transparent Packaging:** It will compile the final application into a single executable that internally orchestrates the graphics engine.

### What Fugo WILL NOT Do

- **Will not compile Go to Dart:** Fugo is not a transpiler.
- **Will not use FFI (Foreign Function Interface) or CGO:** Cross-language communication will avoid shared memory to bypass the complexity of C bindings and guarantee cross-platform portability.
- **Will not use Web technologies:** There are no WebViews, HTML, CSS, or DOM involved.
- **Will not enforce Material Design:** Although it will support standard components, the framework is designed to allow raw, brutalist, or highly customized styles from scratch.

---

## 3. Tech Stack and Architecture

The architecture is divided into two domains that communicate across a strict local network boundary:

| Component                | Technology                  | Purpose                                                                                                                                                       |
| ------------------------ | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **SDK and Logical Core** | Go (1.22+)                  | Provides the declarative API, manages state (Virtual DOM), and handles business logic.                                                                        |
| **Rendering Engine**     | Flutter / Dart              | A generic, precompiled client with no business logic, dedicated exclusively to interpreting and drawing on screen.                                            |
| **Transport Layer**      | gRPC (Bidirectional)        | Maintains the asynchronous, strongly typed data flow. Uses Unix Domain Sockets (UDS) on UNIX systems and TCP `localhost` on Windows for kernel-level latency. |
| **Serialization**        | Protocol Buffers (Protobuf) | Serializes the UI tree into a binary format, eliminating the high computational cost of parsing JSON on every frame.                                          |

---

## 4. Developer Experience (DX)

Fugo is designed to be operated entirely from the terminal, integrating naturally with modern editors (Neovim, VSCode, Goland).

### CLI (Command Line Interface)

The command-line tool will be the center of operations:

- `fugo init <name>`: Generates the base project structure with dependencies ready.
- `fugo run`: Starts the Go server, boots the Flutter engine subprocess, and links both parts instantly.
- `fugo build`: Packages the Go binary alongside the precompiled Flutter client, generating a distribution-ready deliverable (`.exe`, `.app`, or Linux binary).

### Server-Side Hot-Reload

Since the UI is dictated by the state in Go, Fugo will implement a reload mechanism where, upon saving a `.go` file, the server restarts the logic and transmits the new UI tree to the Flutter client without needing to close and reopen the OS window.

---

## 5. Code Experience (The API)

The Go API design will adopt a declarative, functional, and strongly typed pattern, structured as a tree.

### Practical Example: `main.go`

```go
package main

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

func main() {
	// One call starts the gRPC server, spawns the Flutter client, builds the UI
	// once, and runs until the window closes.
	fugo.RunStandalone(fugo.AppOptions{
		Title:  "Fugo Desktop",
		Width:  800,
		Height: 600,
	}, buildUI)
}

func buildUI(ctx *fugo.Context) fg.Widget {
	counter := 0

	// Reactive node: the handler mutates this in place, then calls ctx.Update.
	counterText := fg.Text("0").FontSize(48)

	incrementBtn := fg.Button("Increment").
		BgColor(fg.Hex("#10B981")).
		BorderRadius(4).
		OnClick(func(_ fg.Event) {
			counter++
			counterText.SetText(strconv.Itoa(counter))
			ctx.Update() // mark dirty → diff → patch streamed to Flutter
		})

	// Layout composition. Constructors are prefix-free and return concrete
	// widget types with chainable setters.
	return fg.Container(
		fg.Center(
			fg.Column(
				counterText,
				fg.SizedBox(0, 24),
				incrementBtn,
			),
		),
	).BgColor(fg.Hex("#121212"))
}
```

---

## 6. Performance and Optimization Mechanisms

To ensure Fugo competes with native applications and does not suffer from common bottlenecks in local client-server architectures:

1. **Tree Diffing (Virtual DOM in Go):** Instead of sending the entire UI tree to Flutter on every event, Go will calculate the difference (diff) between the previous state and the new state. Only the changed properties will be sent via Protobuf.
2. **Debounced Asynchronous Events:** High-frequency events (like mouse movement or rapid typing in a text field) will be processed using debouncing and throttling on the client side (Flutter) before flooding the gRPC channel to Go.
3. **Memory and Lifecycle Management:** The Flutter subprocess will be subordinated to the Go binary. If the Go process terminates (due to a `panic` or user termination), the UDS/TCP channel closes and the graphics engine self-destructs immediately to prevent memory leaks or zombie processes.
