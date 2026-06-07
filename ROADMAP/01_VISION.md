# 01 — Visión General

## El problema que Fugo resuelve

### Contexto histórico

El desarrollo de aplicaciones de escritorio con interfaces gráficas modernas ha estado fragmentado durante décadas:

- **Electron (2013)**: Permitió usar tecnologías web para desktop. El costo: ~150MB de RAM mínima, un proceso Chromium por aplicación, y JavaScript como lenguaje de lógica de negocio. Funciona, pero es ineficiente por diseño.
- **Flutter Desktop (2019+)**: Renderizado nativo vía Impeller/Skia, rendimiento 60/120fps, tipografía y layouts de clase mundial. El problema: **obliga a escribir TODO en Dart**, un lenguaje diseñado para frontend, sin el ecosistema de Go para backend, sistemas, y concurrencia.
- **Go + GUI nativa**: Librerías como Fyne, Gio, Wails han intentado llevar GUI a Go. Todas sufren de: renderizado inferior a Flutter, ecosistema de widgets pobre, o dependencia en WebViews.

**El resultado**: Los desarrolladores Go que necesitan GUIs desktop se ven forzados a elegir entre rendimiento pobre (Electron/WebView), ecosistema limitado (Fyne/Gio), o abandonar Go por completo (Dart/Swift/C#).

### La brecha específica

| | Go | Flutter/Dart |
|---|---|---|
| Concurrencia | Goroutines (~2KB, user-space) | Isolates (~2-10MB, message-passing) |
| Ecosistema backend | `database/sql`, gRPC, sistemas de archivos, redes | Paquetes Dart limitados |
| Tipografía/Layouts | Inexistente nativamente | Clase mundial (Impeller, flexbox) |
| Renderizado 60fps | No disponible | Nativo |
| CLI/DevOps | Nativo y maduro | Secundario |

Fugo elimina esta disyuntiva: Go para TODO el backend y la lógica, Flutter como motor de renderizado puro.

---

## Qué es Fugo — definición precisa

Fugo es un framework **Server-Driven UI (SDUI)** local para aplicaciones desktop:

```
┌──────────────────────┐       IPC (UDS)       ┌──────────────────────┐
│   Go Process          │◄══════════════════════►│   Flutter Process     │
│                       │   FlatBuffers + gRPC   │                       │
│  ┌─────────────────┐  │                        │  ┌─────────────────┐  │
│  │ Lógica de negocio│  │                        │  │ Widget Registry  │  │
│  │ State management │  │   Widget Tree Diff     │  │ Render Pipeline  │  │
│  │ Virtual DOM      │──┼───────────────────────►│  │ Event Handler    │  │
│  │ gRPC Server      │  │                        │  │ gRPC Client      │  │
│  └─────────────────┘  │   User Events           │  └─────────────────┘  │
│                       │◄───────────────────────│                       │
└──────────────────────┘                        └──────────────────────┘
```

**Go es la fuente absoluta de verdad.** Flutter es un terminal de renderizado sin lógica de negocio. La comunicación es local (sub-milisegundo) vía Unix Domain Sockets, no por red.

---

## Lo que Fugo NO es

| No es | Por qué |
|-------|---------|
| Un transpilador Go→Dart | No compila Go a Dart. Go ejecuta Go, Flutter ejecuta Dart precompilado. |
| Un puente FFI/CGO | No usa C bindings ni memoria compartida. La comunicación es por IPC con mensajes tipados. |
| Una WebView glorificada | No hay HTML, CSS, DOM ni navegador embebido. Renderizado nativo vía Impeller. |
| Un clon de Material Design | No impone sistema de diseño. Provee primitivas atómicas para que el desarrollador construya su propio estilo. |
| Una plataforma de plugins | No es un ecosistema de extensiones. Es una herramienta autocontenida. |

---

## Inspiraciones y linaje técnico

### Flet (Python + Flutter) — [flet.dev](https://flet.dev)

**Lo que tomamos**: El concepto de SDUI con Flutter como motor de renderizado genérico. Flet demostró que es viable separar la lógica (Python) del renderizado (Flutter).

**Lo que mejoramos**: Flet usa JSON sobre WebSockets — lento, verboso, sin tipado fuerte. Fugo usa FlatBuffers sobre UDS — zero-copy, tipado, sub-10µs de latencia. Flet sufre en concurrencia por el GIL de Python; Fugo aprovecha goroutines nativas de Go.

### Phoenix LiveView (Elixir) — [hexdocs.pm/phoenix_live_view](https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.html)

**Lo que tomamos**: El modelo de "estado vive en el servidor, solo se envían diffs al cliente". LiveView calcula diffs del DOM y envía parches mínimos por WebSocket.

**Lo que mejoramos**: LiveView está limitado a HTML/CSS. Fugo aplica el mismo principio a widgets nativos Flutter. La latencia de red no es problema porque la comunicación es local (UDS, no WebSocket por internet).

### Airbnb Epoxy — [github.com/airbnb/epoxy](https://github.com/airbnb/epoxy)

**Lo que tomamos**: El patrón de "cada widget tiene un modelo tipado con ID estable, y el sistema diffea por ID". Epoxy usa modelos con hash-based change detection para actualizar solo lo que cambió.

**Lo que adaptamos**: Epoxy es Android-only y no serializa por red. Fugo toma el concepto de modelo tipado por widget y lo extiende a comunicación cross-process con FlatBuffers.

### The Elm Architecture (MVU) — [guide.elm-lang.org/architecture](https://guide.elm-lang.org/architecture/)

**Lo que tomamos**: Flujo de datos unidireccional. La UI es una función pura del estado. Eventos disparan mutaciones de estado, que producen una nueva UI. Sin efectos secundarios en el renderizado.

**Lo que adaptamos**: No usamos el runtime de Elm ni su sistema de efectos. Go maneja el estado con structs y goroutines. El Model-View-Update emerge naturalmente del patrón `struct → Render() → UI tree`.

---

## Principios de diseño inamovibles

1. **Go es el source of truth absoluto.** Toda la lógica, estado, routing y validación vive en Go. Flutter es estrictamente un terminal de renderizado.

2. **Sin shared state.** No hay estado compartido entre Go y Flutter. Cada lado tiene su memoria aislada. La comunicación es por mensajes tipados.

3. **Zero-copy donde sea posible.** FlatBuffers permite leer el árbol de widgets sin deserializar. Las diffs se comparan como slices de bytes (`bytes.Equal`).

4. **Opinionated en state management, unopinionated en diseño visual.** Fugo impone UN solo patrón de estado (component model). Pero no impone Material Design, Cupertino, ni ningún sistema visual. El desarrollador compone sus propios design systems con primitivas atómicas.

5. **Rendimiento de primer orden.** No es una optimización tardía. Cada decisión de arquitectura (UDS sobre TCP, FlatBuffers sobre JSON, array plano sobre árbol de punteros) se toma por rendimiento.

6. **Operable desde terminal.** `fugo init`, `fugo run`, `fugo build`. Sin GUIs para configurar, sin asistentes gráficos. La CLI es el centro de operaciones.

---

## Por qué esto funciona (y otros SDUI fallaron)

### Por qué los SDUI remotos fallan

Shopify, LinkedIn (Voyager), Etsy — todos intentaron SDUI sobre redes móviles. El problema fundamental: **latencia de red** (50-200ms por round-trip en 4G). Cada interacción del usuario requiere un viaje al servidor, haciendo la UI sentir "lenta".

Fugo **elimina este problema por diseño**: la comunicación es local, sobre Unix Domain Sockets, con latencias de 5-10µs (microsegundos, no milisegundos). Esto es 10,000x más rápido que una red móvil.

### Por qué los SDUI locales son viables

- **UDS latencia**: 5-10µs por mensaje (vs 50-200ms en 4G)
- **UDS throughput**: 5-10 GB/s (vs ~10-50 MB/s en WiFi)
- **Sin pérdida de paquetes**: local, no hay congestión de red
- **Sin serialización costosa**: FlatBuffers zero-copy vs JSON parse
- **Sin overhead de seguridad**: no hay TLS, autenticación, ni cifrado entre procesos locales

El "problema del SDUI" siempre fue la red. Al eliminarla, el modelo se vuelve no solo viable, sino superior: separación limpia de lógica y presentación, sin el costo de latencia.

---

## Referencias

- Flet framework: <https://flet.dev/docs/>
- Phoenix LiveView: <https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.html>
- Airbnb Epoxy: <https://github.com/airbnb/epoxy>
- Elm Architecture: <https://guide.elm-lang.org/architecture/>
- Shopify SDUI (blog): <https://shopify.engineering/>
- LinkedIn Voyager: <https://engineering.linkedin.com/blog/2017/05/building-a-server-driven-ui-framework>
- Flutter architectural overview: <https://docs.flutter.dev/resources/architectural-overview>
- SPEC.md del proyecto: `../SPEC.md`
