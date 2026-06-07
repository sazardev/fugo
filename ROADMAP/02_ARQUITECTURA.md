# 02 — Arquitectura del Sistema

## Visión general

Fugo es un sistema de dos procesos que se comunican localmente mediante IPC. La arquitectura se divide en **3 capas lógicas** distribuidas en **2 procesos del sistema operativo**:

```
┌─────────────────────────────────────────────────────────┐
│                   Proceso Go (Host)                      │
│                                                          │
│  ┌──────────────────────────────────────────────────┐   │
│  │              API Pública (Go SDK)                  │   │
│  │  ui.Container, ui.Text, ui.Button, ui.Row, ...    │   │
│  │  style.Font(), style.BgColor(), ...               │   │
│  │  app.Run(), ctx.Update(), ctx.NavigateTo()        │   │
│  └──────────────────┬───────────────────────────────┘   │
│                     │                                    │
│  ┌──────────────────▼───────────────────────────────┐   │
│  │            Core Engine (Go)                       │   │
│  │  ┌─────────────┐  ┌──────────┐  ┌─────────────┐  │   │
│  │  │ Virtual DOM  │  │  Diffing  │  │  Scheduler   │  │   │
│  │  │ (flat array) │  │  Engine   │  │  (debounce)  │  │   │
│  │  └─────────────┘  └──────────┘  └─────────────┘  │   │
│  └──────────────────┬───────────────────────────────┘   │
│                     │                                    │
│  ┌──────────────────▼───────────────────────────────┐   │
│  │            Transport Layer (Go)                   │   │
│  │  ┌─────────────────────┐  ┌────────────────────┐  │   │
│  │  │ FlatBuffers Codec    │  │ gRPC Server         │  │   │
│  │  │ (marshal/unmarshal) │  │ (bidirectional)     │  │   │
│  │  └─────────────────────┘  └────────────────────┘  │   │
│  │  ┌─────────────────────────────────────────────┐  │   │
│  │  │ Process Supervisor (os/exec + signals)      │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  └──────────────────┬───────────────────────────────┘   │
│                     │ UDS / localhost TCP                │
└─────────────────────┼───────────────────────────────────┘
                      │
┌─────────────────────┼───────────────────────────────────┐
│                     │     Proceso Flutter (Client)        │
│  ┌──────────────────▼───────────────────────────────┐   │
│  │            Transport Layer (Dart)                 │   │
│  │  ┌─────────────────────┐  ┌────────────────────┐  │   │
│  │  │ FlatBuffers Codec    │  │ gRPC Client         │  │   │
│  │  │ (read, zero-copy)   │  │ (bidirectional)     │  │   │
│  │  └─────────────────────┘  └────────────────────┘  │   │
│  └──────────────────┬───────────────────────────────┘   │
│                     │                                    │
│  ┌──────────────────▼───────────────────────────────┐   │
│  │            Rendering Engine (Dart)                │   │
│  │  ┌─────────────────────┐  ┌────────────────────┐  │   │
│  │  │ Widget Registry      │  │ Event Debouncer     │  │   │
│  │  │ (Factory pattern)   │  │ (throttle events)   │  │   │
│  │  └─────────────────────┘  └────────────────────┘  │   │
│  │  ┌─────────────────────────────────────────────┐  │   │
│  │  │ Widget Tree Builder (recursive)              │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  └──────────────────┬───────────────────────────────┘   │
│                     │                                    │
│  ┌──────────────────▼───────────────────────────────┐   │
│  │            Platform Layer                         │   │
│  │  ┌─────────────────┐  ┌────────────────────────┐  │   │
│  │  │ Flutter Engine   │  │ Window Manager          │  │   │
│  │  │ (Impeller/Skia)  │  │ (window_manager pkg)   │  │   │
│  │  └─────────────────┘  └────────────────────────┘  │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

---

## Componentes y responsabilidades

### 1. Go SDK (API Pública)

**Ubicación**: Paquete `github.com/sazardev/fugo`

**Responsabilidad**: Proporcionar la API declarativa que el desarrollador usa para definir la UI. Es la única superficie que el desarrollador toca.

```go
app := fugo.NewApp(fugo.AppOptions{Title: "Mi App", Width: 800, Height: 600})
app.Run(func(ctx *fugo.Context) ui.Widget {
    return ui.Container(
        ui.Text("Hola Fugo").FontSize(24),
    ).BgColor("#1E1E1E")
})
```

**Subpaquetes**:
- `fugo` — Inicialización del framework, App, Context, ciclo de vida
- `fugo/ui` — Widgets (Container, Text, Button, Row, Column, etc.)
- `fugo/style` — Primitivas de estilo (Font, Color, Border, Shadow, Gradient)

[Ver 07_API_GO.md para especificación completa]

### 2. Core Engine (Go)

**Responsabilidad**: Motor interno que convierte la API declarativa en diffs serializables.

**Subcomponentes**:

| Componente | Función | Estructura de datos |
|-----------|---------|-------------------|
| Virtual DOM | Representación plana del árbol de widgets | `[]VNode` (array, no árbol de punteros) |
| Diffing Engine | Compara old VDOM vs new VDOM, produce parches | Algoritmo O(n) por ID |
| Scheduler | Controla cuándo se envían actualizaciones a Flutter | Debounce 16ms (alineado a frame) |
| Reconciler | Aplica diff y dispara serialización | Coordina Diffing → FlatBuffers → gRPC |

[Ver 03_CORE_SDK.md para especificación completa]

### 3. Transport Layer (Go + Dart)

**Responsabilidad**: Comunicación bidireccional entre procesos con tipado fuerte y zero-copy donde sea posible.

**Stack de transporte**:

```
Capa de aplicación:  Widget Tree / Events  (FlatBuffers)
Capa de RPC:         Streaming RPC         (gRPC)
Capa de transporte:  HTTP/2                (gRPC built-in)
Capa de red:         Unix Domain Socket    (Linux/macOS)
                     TCP localhost         (Windows fallback)
```

**Canales**:
- **Go → Flutter**: `RenderPayload` — diffs del árbol de widgets (FlatBuffer)
- **Flutter → Go**: `EventPayload` — eventos de usuario (click, hover, input)

[Ver 05_TRANSPORTE.md para especificación completa]

### 4. Rendering Engine (Dart/Flutter)

**Responsabilidad**: Recibir descripciones de widgets, construirlos en el árbol de Flutter, renderizarlos, y enviar eventos de vuelta.

**Subcomponentes**:

| Componente | Función |
|-----------|---------|
| Widget Registry | Mapea tipos de widget (string → FugoWidgetBuilder) |
| Widget Tree Builder | Construye recursivamente el árbol de Flutter desde FlatBuffer |
| Event Debouncer | Filtra eventos de alta frecuencia (mouse move, keystrokes) antes de enviar a Go |
| Reconnection Handler | Detecta caída de conexión UDS y reconecta automáticamente |

[Ver 04_FLUTTER_CLIENT.md para especificación completa]

### 5. Process Supervisor (Go)

**Responsabilidad**: Gestionar el ciclo de vida del proceso Flutter como subproceso de Go.

```
Go Host (padre)
  ├── spawn: os/exec → Flutter binary
  ├── heartbeat: ping/pong cada 1000ms
  ├── signals: propagate SIGTERM, handle exit
  └── cleanup: remove UDS socket file on exit
```

[Ver 08_DESKTOP.md para especificación completa]

---

## Flujo de datos (ciclo de vida de una interacción)

### Ejemplo: click en un botón que incrementa un contador

```
PASO 1: Usuario hace click en el botón "Incrementar"
  Flutter: GestureDetector.onTap → Event Debouncer → FlatBuffer encode → gRPC stream.Send()

PASO 2: Evento llega a Go (~10µs después por UDS)
  Go: gRPC stream.Recv() → FlatBuffer decode → dispatch a handler
  Go: handler ejecuta: counter++; counterText.SetText("1"); ctx.Update()

PASO 3: Go recalcula UI y diffea
  Go: app.Run() closure se re-ejecuta → genera nuevo Virtual DOM
  Go: Diffing Engine compara old VDOM vs new VDOM
  Go: Produce patch: UPDATE node_counter_text { text: "1" }

PASO 4: Diff se envía a Flutter
  Go: FlatBuffer encode patch → gRPC stream.Send()

PASO 5: Flutter aplica diff (~10µs después)
  Flutter: gRPC stream.Recv() → FlatBuffer decode → Widget Registry
  Flutter: Actualiza Text widget con nuevo valor → Impeller re-renderiza frame

TOTAL: ~30-100µs desde click hasta frame renderizado
```

---

## Decisiones arquitectónicas fundamentales

### Decisión 1: Dos procesos, no un solo proceso con FFI

**Alternativas consideradas**:
- **FFI/CGO**: Llamar al motor Flutter desde Go vía C bindings
- **WASI/WASM**: Ejecutar la lógica Go dentro del runtime de Dart

**Decisión**: Dos procesos separados comunicándose por IPC.

**Justificación**:
- FFI/CGO rompe la portabilidad (CGO no cross-compila bien) y añade complejidad de manejo de memoria entre GC de Go y GC de Dart.
- Dos procesos aislados significa que un panic en Go no corrompe el motor Flutter, y viceversa.
- La latencia de IPC local (5-10µs) es insignificante comparada con el presupuesto de frame (16,667µs).
- Permite desarrollar, testear y debugear cada lado independientemente.

**Referencia**: `SPEC.md:23-25` — "Will not use FFI or CGO: Cross-language communication will avoid shared memory to bypass the complexity of C bindings and guarantee cross-platform portability."

### Decisión 2: gRPC + FlatBuffers, no gRPC + Protobuf puro

**Alternativas consideradas**:
- gRPC con Protobuf estándar (google.golang.org/protobuf)
- WebSockets + JSON (como Flet)
- Raw UDS + MessagePack

**Decisión**: gRPC como capa RPC (control, ciclo de vida) + FlatBuffers como formato de datos para el árbol de widgets.

**Justificación**:
- gRPC provee streaming bidireccional, health checking, manejo de errores, y reconexión out-of-the-box.
- Protobuf estándar requiere heap allocation en cada deserialización. FlatBuffers permite zero-copy reads.
- FlatBuffers tiene soporte oficial en Dart y Go.
- Para la capa de control (health, shutdown), Protobuf es suficiente y más simple.

[Ver 05_TRANSPORTE.md:§3-4 para benchmarks y comparativas]

### Decisión 3: Virtual DOM con array plano, no árbol de punteros

**Alternativas consideradas**:
- Árbol de structs con punteros (tradicional React/Vue)
- Representación basada en strings (HTML templates como LiveView)

**Decisión**: Array plano (`[]VNode`) indexado por ID.

**Justificación**:
- Mejor localidad de caché: los nodos están contiguos en memoria.
- Comparación O(1) por ID sin seguir punteros.
- La diffeo recorre el array secuencialmente (CPU branch predictor friendly).
- FlatBuffers serializa arrays de structs de forma natural.

[Ver 03_CORE_SDK.md:§3 para implementación]

### Decisión 4: Debouncing en Flutter, no en Go

**Alternativas consideradas**:
- Enviar todos los eventos raw a Go y que Go los filtre
- Filtrar en ambos lados

**Decisión**: Debouncing/throttling en el cliente Flutter antes de enviar a Go.

**Justificación**:
- Reduce tráfico en el canal gRPC (los eventos de mouse pueden ser 1000+/segundo).
- Go recibe eventos ya limpios y significativos.
- Flutter tiene acceso al `SchedulerBinding` para alinear eventos al frame.

**Referencia**: Flutter performance best practices — <https://docs.flutter.dev/perf/best-practices#build-and-display-frames-in-16ms>

### Decisión 5: window_manager para gestión de ventanas

**Alternativas consideradas**:
- bitsdojo_window
- Flutter nativo (PlatformMenuBar experimental)
- Implementación propia en C++ por plataforma

**Decisión**: `window_manager` (leanflutter) como base.

**Justificación**:
- Soporta Linux (GTK), Windows (Win32), macOS (Cocoa) con API unificada.
- Permite frameless windows y custom title bars (necesario para el enfoque "blank slate" de Fugo).
- Comunidad activa (score 88.88 en pub.dev).

[Ver 08_DESKTOP.md:§2 para detalles de integración]

---

## Modelo de concurrencia

```
Go Process                           Flutter Process
═══════════                          ═══════════════
Main Goroutine                       Main Isolate (UI thread)
  ├── App.Run() event loop             ├── Widget build/layout/paint
  ├── gRPC server (goroutine)          ├── gRPC client (Background Isolate)
  ├── Diffing (goroutine)              │     ├── FlatBuffer decode
  ├── Heartbeat (goroutine)            │     ├── Event encode
  └── Signal handler (goroutine)       │     └── Reconnection logic
                                       └── Platform channels
```

**Principio clave**: Go usa goroutines para concurrencia (memoria compartida). Flutter usa isolates (memoria aislada, message-passing). La comunicación entre procesos es el único punto de sincronización.

---

## Estructura del repositorio

```
fugo/
├── cmd/
│   └── fugo/           # CLI binary (fugo init, run, build)
├── fugo.go             # Package fugo: App, Context, Options
├── ui/
│   ├── widget.go       # Interface Widget + builder pattern
│   ├── container.go    # ui.Container
│   ├── text.go         # ui.Text
│   ├── button.go       # ui.Button
│   ├── layout.go       # ui.Row, ui.Column, ui.Stack, ui.Expanded
│   ├── input.go        # ui.TextField, ui.Checkbox, ui.Slider
│   ├── list.go         # ui.ListView, ui.GridView
│   ├── animation.go    # ui.AnimatedContainer, ui.AnimatedOpacity
│   └── router.go       # ui.Router, ui.Route
├── style/
│   ├── style.go        # Style struct + chaining
│   ├── color.go        # Color parsing, hex, rgba
│   ├── font.go         # Font families, weights
│   ├── border.go       # Borders, radius
│   └── shadow.go       # Box shadows
├── engine/
│   ├── vdom.go         # Virtual DOM (flat array)
│   ├── differ.go       # Diffing algorithm
│   ├── reconciler.go   # Apply diff + trigger render
│   └── scheduler.go    # Debounce + frame alignment
├── transport/
│   ├── proto/           # FlatBuffers schema (.fbs)
│   ├── codec.go         # FlatBuffer marshal/unmarshal
│   └── server.go        # gRPC server setup
├── supervisor/
│   ├── process.go       # os/exec Flutter child
│   ├── heartbeat.go     # Health checking
│   └── signals.go       # OS signal handling
├── flutter_client/      # Proyecto Flutter independiente
│   ├── lib/
│   │   ├── main.dart
│   │   ├── registry/    # Widget registry
│   │   ├── transport/   # gRPC client, FlatBuffer codec
│   │   └── events/      # Event debouncer
│   └── pubspec.yaml
├── docs/                # Documentación técnica
├── ROADMAP/             # Este directorio
├── SPEC.md
├── Makefile
└── go.mod
```

---

## Referencias

- Flet architecture: <https://flet.dev/docs/>
- Flutter embedder API: <https://github.com/flutter/engine/blob/main/shell/platform/embedder/embedder.h>
- gRPC health checking: <https://grpc.io/docs/guides/health-checking/>
- FlatBuffers Dart: <https://pub.dev/packages/flat_buffers>
- FlatBuffers Go: <https://github.com/google/flatbuffers/go>
- window_manager: <https://pub.dev/packages/window_manager>
