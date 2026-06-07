# ANNEX.md: Fugo - Technical Deep Dive & Architectural Relations

## 1. Architectural Inspirations & Paradigms

Fugo does not exist in a vacuum; it synthesizes proven concepts from various ecosystems into a single, high-performance desktop framework.

- **Flet / Flet.dev (Python + Flutter):** Fugo inherits the core concept of Server-Driven UI (SDUI) using Flutter as a generic rendering engine. However, Fugo diverges by utilizing Go's native concurrency (Goroutines) and strictly binary protocols (Protobuf/gRPC) instead of JSON/WebSockets, aiming for zero-allocation performance where possible.
- **Phoenix LiveView (Elixir):** The concept of keeping the state entirely on the server and pushing minimal diffs (deltas) over a persistent connection. Fugo applies this web concept to local Inter-Process Communication (IPC).
- **The Elm Architecture (MVU - Model, View, Update):** Fugo’s API encourages a unidirectional data flow. The UI is a pure function of the Go state. Events trigger state mutations in Go, which in turn generate a new UI tree to be reconciled.

---

## 2. The Protobuf Contract (The IPC Bridge)

The bottleneck in any SDUI framework is serialization. To achieve 60+ FPS, Fugo uses Protocol Buffers. The contract between Go and Flutter must be strictly defined to avoid reflection overhead.

### Conceptual `.proto` Definition

```protobuf
syntax = "proto3";
package fugo.v1;

// The payload sent from Go to Flutter
message RenderPayload {
  // Can be a full tree or a patch (diff)
  enum PayloadType {
    FULL_TREE = 0;
    PATCH = 1;
  }
  PayloadType type = 1;
  repeated Node nodes = 2;
}

// A generic UI Node
message Node {
  string id = 1;
  string widget_type = 2; // e.g., "Text", "Container", "Button"

  // Encoded properties to avoid complex nested objects
  map<string, bytes> properties = 3;
  repeated string children_ids = 4;
}

// The payload sent from Flutter to Go
message EventPayload {
  string node_id = 1;
  string event_type = 2; // e.g., "onClick", "onHover", "onChange"
  bytes event_data = 3;  // e.g., text field input value
}

```

---

## 3. The Virtual DOM & Diffing Engine (Go-Side)

To prevent overwhelming the gRPC channel and the Flutter rendering engine, Go must never send the entire UI tree unless absolutely necessary (e.g., the initial render).

### The Diffing Lifecycle

1. **State Mutation:** An event triggers a state change in Go (e.g., `counter++`).
2. **Re-evaluation:** The Go `Run` closure executes again, generating a new Virtual Tree in memory.
3. **Reconciliation (O(n) complexity):** Go compares the New Tree against the Old Tree.
4. **Patch Generation:** Go creates a list of operations:

- `UPDATE node_1 (text: "1")`
- `INSERT node_2 (parent: col_1)`
- `DELETE node_3`

5. **Dispatch:** Only the patches are serialized into Protobuf and sent to Flutter.

---

## 4. UI Rendering Mechanics (Flutter-Side)

The Flutter client is fundamentally a "dumb" terminal. It contains a global `Map<String, Widget>` and a recursive builder.

### The Widget Registry Pattern

Instead of a massive, unmaintainable `switch` statement, the Flutter client will use a Registry Pattern. Each supported Widget type will have a dedicated deserializer.

```dart
// Conceptual Dart Implementation
abstract class FugoWidgetBuilder {
  Widget build(BuildContext context, Map<String, dynamic> properties, List<Widget> children);
}

class FugoTextBuilder implements FugoWidgetBuilder {
  @override
  Widget build(BuildContext context, Map<String, dynamic> properties, List<Widget> children) {
    return Text(
      properties['value'] ?? '',
      style: TextStyle(
        fontSize: properties['font_size']?.toDouble(),
        color: hexToColor(properties['color']),
        fontFamily: properties['font_family'],
      ),
    );
  }
}

```

---

## 5. Design System Freedom (The "Raw Primitive" Approach)

Most frameworks force developers into Material Design or Cupertino. Fugo takes inspiration from the **Brutalist** and **Nothing UI** design trends.

Fugo will achieve this by exposing "Atoms" rather than "Molecules".

- **Instead of `ui.MaterialCard**`, Fugo provides `ui.Container`.
- Developers build their own abstractions in Go. If a developer wants a harsh, monochrome UI with _General Sans_, they define a Go struct that wraps `ui.Container` with strict 2px black borders and zero drop-shadows.

```go
// Example of a developer creating their own Brutalist Design System on top of Fugo
func BrutalistButton(label string, onClick func(ui.Event)) ui.Widget {
    return ui.Container(
        ui.Text(label).Font("General Sans").Weight(900).Color("#000000"),
    ).
    Border(2, "#000000").
    BgColor("#FFFFFF").
    PaddingXY(24, 12).
    OnClick(onClick)
}

```

---

## 6. Subprocess Lifecycle Management (The Go Supervisor)

The user should never have to manage the Flutter process manually. Go acts as the orchestrator.

1. **Execution:** `go run main.go` is executed.
2. **Port Binding:** Go finds an available UDS (Unix Domain Socket) or a random TCP port.
3. **Forking:** Go uses `os/exec` to launch the Flutter binary hidden inside the compiled package, passing the port as an environment variable (`FUGO_PORT=45021`).
4. **Heartbeat:** Go and Flutter exchange Ping/Pong packets every 1000ms.
5. **Termination Gracefully:** If the Flutter window is closed by the user, Flutter sends a `SIGTERM` to Go. If Go panics or is killed via terminal (`Ctrl+C`), the pipe is broken, and Flutter immediately calls `exit(0)`.
