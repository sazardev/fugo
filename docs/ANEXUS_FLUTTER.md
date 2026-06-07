# ANNEX-FLUTTER.md: Flutter Engine Scope & API Mapping Surface

## 1. Overview: The API Mapping Strategy

Fugo does not transpile Go to Dart. Instead, it creates a **1:1 mapping** between Go structs and Flutter Widgets through the Protobuf serialization contract. To provide the full power of Flutter to the Go developer, Fugo must meticulously map a vast subset of the Flutter API.

This document outlines the exhaustive scope of Flutter features, widgets, properties, and system-level capabilities that Fugo will replicate, wrap, and manage.

---

## 2. Structural & Layout Architecture

Flutter’s layout algorithm is constraints-based. Fugo will implement the core layout primitives, allowing the Go developer to compose complex interfaces.

| Flutter Concept        | Fugo Go Equivalent          | Scope & Managed Properties                                                                                                                                |
| ---------------------- | --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Container`            | `ui.Container`              | Padding, Margin, Alignment, Constraints (min/max width/height), Color, Decoration.                                                                        |
| `Row` & `Column`       | `ui.Row`, `ui.Column`       | `MainAxisAlignment`, `CrossAxisAlignment`, `MainAxisSize`, spacing gaps.                                                                                  |
| `Stack` & `Positioned` | `ui.Stack`, `ui.Positioned` | Z-index rendering, absolute positioning (`Top`, `Right`, `Bottom`, `Left`).                                                                               |
| `Flex` & `Expanded`    | `ui.Flex`, `ui.Expanded`    | Flex factors, flexible space distribution within Rows/Columns.                                                                                            |
| `Wrap`                 | `ui.Wrap`                   | Direction, spacing, run spacing, alignment (for responsive reflowing layouts).                                                                            |
| `ListView`             | `ui.ListView`               | Scroll direction, physics. **Crucial:** Mapped as a virtualized list. Go sends chunks of data dynamically (Pagination via gRPC) to avoid memory overflow. |
| `GridView`             | `ui.GridView`               | Cross-axis count, aspect ratio, main/cross axis spacing.                                                                                                  |

---

## 3. Visual Primitives & Display Nodes

These are the leaf nodes of the UI tree. They do not have children and are solely responsible for painting content on the screen.

- **`ui.Text` (`Text` / `RichText`):**
- _Managed Properties:_ String value, `FontSize`, `FontFamily`, `FontWeight`, `Color`, `LetterSpacing`, `LineHeight`, `TextAlign`, `Overflow` (ellipsis, fade).

- **`ui.Image` (`Image.network`, `Image.asset`, `Image.memory`):**
- _Managed Properties:_ Source URL or Base64 byte array, `BoxFit` (cover, contain, fill), `ColorFilter`, width/height.

- **`ui.Icon` (`Icon`):**
- _Managed Properties:_ SVG path string or material icon codepoint, size, color.

- **`ui.Divider` (`Divider` / `VerticalDivider`):**
- _Managed Properties:_ Thickness, color, indent, end-indent.

---

## 4. Input & Form Controls (Two-Way Binding)

Input controls are the most complex elements to map because they require managing asynchronous state discrepancies between the Flutter UI thread and the Go backend.

- **`ui.TextField` (`TextField` / `TextFormField`):**
- _Scope:_ Value, placeholder/hint, obscure text (passwords), max lines, keyboard type.
- _Event Management:_ Keystrokes must be throttled/debounced on the Flutter side (e.g., send `onChange` event to Go every 300ms of inactivity) to prevent overloading the gRPC channel.

- **`ui.Checkbox` & `ui.Switch` (`Checkbox`, `Switch`):**
- _Scope:_ Boolean state, active color, inactive track color.

- **`ui.Slider` (`Slider`):**
- _Scope:_ Min, max, divisions, current value. Emits `onChangeEnd` to Go to trigger heavy calculations.

- **`ui.Dropdown` (`DropdownButton`):**
- _Scope:_ List of string/ID key-value pairs, selected ID, dropdown menu styling.

---

## 5. Styling & Decoration (The Painting Layer)

Fugo will expose the power of Flutter's `BoxDecoration` to allow raw, customized styling.

- **Borders:** Solid, dashed, thickness, directional borders (only bottom, only top).
- **Border Radius:** Circular, elliptical, individual corners (e.g., top-left only).
- **Shadows (`BoxShadow`):** Color, X/Y offset, blur radius, spread radius. Enables Neumorphic or floating designs.
- **Gradients (`LinearGradient`, `RadialGradient`):** Array of colors, stops, begin/end alignments.
- **Clipping (`ClipRRect`, `ClipOval`):** Hard clipping of child widgets (e.g., circular profile pictures).

---

## 6. Gestures & Event Handlers

Flutter’s `GestureDetector` will be wrapped into a unified event listening system inside Fugo.

| Flutter Event           | Fugo Go Event Handler             | Fugo Payload Sent to Go                                                           |
| ----------------------- | --------------------------------- | --------------------------------------------------------------------------------- |
| `onTap`                 | `OnClick(func(e ui.Event))`       | Node ID, Timestamp.                                                               |
| `onDoubleTap`           | `OnDoubleClick(func(e ui.Event))` | Node ID, Timestamp.                                                               |
| `onLongPress`           | `OnLongPress(func(e ui.Event))`   | Node ID, Timestamp.                                                               |
| `onHover` (MouseRegion) | `OnHover(func(e ui.HoverEvent))`  | Hover state (true/false), X/Y coordinates.                                        |
| `onPanUpdate`           | `OnDrag(func(e ui.DragEvent))`    | Delta X, Delta Y, global position (useful for custom sliders or canvas dragging). |

---

## 7. Motion & Implicit Animations

Fugo avoids complex `AnimationController` synchronization across the gRPC bridge. Instead, it relies entirely on Flutter's **Implicit Animations**.

- **`ui.AnimatedContainer`:** Replaces standard Containers. If Go changes a property (e.g., width from 100 to 200), Flutter automatically interpolates the layout change.
- _Managed Properties:_ `Duration` (milliseconds), `Curve` (easeIn, easeOut, bounce, elastic).

- **`ui.AnimatedOpacity`:** Fades elements in and out based on Go state toggles.
- **`ui.AnimatedPositioned`:** Smoothly moves widgets inside a Stack.

---

## 8. Desktop & System Integration (Window Management)

Since Fugo targets Desktop applications, the Flutter client must integrate deeply with OS APIs (typically via packages like `window_manager` or `bitsdojo_window`), exposing these capabilities to Go.

- **Window Controls:** \* Functions exposed in Go: `app.Window().Minimize()`, `Maximize()`, `Close()`, `SetTitle()`, `SetSize(w, h)`, `Center()`.
- **Frameless / Borderless Windows:** \* Capability to hide the native OS title bar and draw custom title bars in Fugo. Requires exposing an `ui.WindowDragArea` widget that maps to desktop window dragging capabilities.
- **System Dialogs:**
- Native file pickers (Open/Save dialogs) managed by Flutter plugins but triggered and resolved via Go channels.

- **Clipboard Management:**
- Read/Write to OS clipboard directly from Go.

---

## 9. Deliberate Exclusions (What will NOT be mapped)

To maintain performance and framework stability, certain Flutter features are explicitly excluded from the Fugo API mapping:

1. **`CustomPaint` & Raw Canvas API:** Drawing pixel-by-pixel paths from Go would require sending thousands of draw commands per frame over gRPC, destroying performance. Canvas drawing is not supported.
2. **Synchronous UI Callbacks:** Any Flutter feature that requires a _synchronous_ return value from a callback cannot be supported, as gRPC communication is inherently asynchronous.
3. **Complex Custom Slivers:** While standard `ListView` and `GridView` are supported, highly bespoke sliver scrolling behaviors that require frame-by-frame offset calculations will be omitted.
