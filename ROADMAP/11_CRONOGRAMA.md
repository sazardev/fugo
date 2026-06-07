# 11 — Cronograma y Estimaciones

## Nota sobre las estimaciones

Los tiempos están expresados en **semanas de trabajo concentrado** (40h/semana) para un desarrollador con experiencia en Go y Flutter. No son meses-hombre diluidos ni incluyen buffer de gestión.

El orden de las fases está determinado por dependencias técnicas. Algunas fases pueden solaparse parcialmente.

---

## Fases y dependencias

```
Fase A: Transporte        Fase B: Core SDK
(05_TRANSPORTE.md)        (03_CORE_SDK.md)
     │                          │
     └──────────┬───────────────┘
                │
         Fase C: Flutter Client
         (04_FLUTTER_CLIENT.md)
                │
     ┌──────────┼──────────┐
     │          │          │
Fase D:     Fase E:    Fase F:
API Go      Desktop    CLI
(07_API)    (08_DESK)  (06_CLI)
     │          │          │
     └──────────┼──────────┘
                │
         Fase G: Rendimiento
         (09_RENDIMIENTO.md)
```

---

## Fase A: Transporte — 14 semanas

**Archivo**: [05_TRANSPORTE.md]

| Semana | Tarea |
|--------|-------|
| 1-2 | Esquema FlatBuffer (.fbs) — todos los tipos de widget y propiedades |
| 3-4 | Codec FlatBuffer Go (marshal/unmarshal zero-allocation) |
| 5-6 | Codec FlatBuffer Dart (decode zero-copy) |
| 7 | Definición .proto + codegen Go/Dart |
| 8-9 | gRPC Server Go (con vtprotobuf codec) |
| 10 | gRPC Client Dart |
| 11 | UDS setup Linux + macOS |
| 12 | TCP fallback Windows |
| 13 | Health checking + keepalive |
| 14 | Integración y tests E2E |

**Entregable**: Go y Flutter se comunican bidireccionalmente. Latencia <200µs.

---

## Fase B: Core SDK (Go Engine) — 14 semanas

**Archivo**: [03_CORE_SDK.md]

| Semana | Tarea |
|--------|-------|
| 1 | Estructura VDOM (VNode, array plano, index) |
| 2-4 | Algoritmo de Diffing (CREATE, UPDATE, DELETE, REPLACE, REORDER) |
| 5-6 | Patch Protocol + serialización FlatBuffer |
| 7-8 | Reconciler (ciclo VDOM → Diff → Patch) |
| 9 | Scheduler (debounce 16ms) |
| 10-11 | gRPC Server (RenderStream bidireccional) |
| 12-13 | FlatBuffers codec integración (Go side) |
| 14 | Tests y benchmarks |

**Entregable**: Core Engine funcional con tests de diffing, benchmarks de <100µs para 1000 nodos.

---

## Fase C: Flutter Client — 13 semanas

**Archivo**: [04_FLUTTER_CLIENT.md]

**Depende de**: Fase A (transporte funcional), Fase B (protocolo de patches definido)

| Semana | Tarea |
|--------|-------|
| 1 | Estructura del proyecto Flutter + pubspec |
| 2-3 | Background Isolate + gRPC Client |
| 4-5 | FlatBuffer Decoder (Dart side, zero-copy) |
| 6-9 | Widget Registry + Builders (15 widgets base) |
| 10-11 | Widget Tree Builder + Patch Application |
| 12 | Event Debouncer (throttle + debounce) |
| 13 | Reconnection + Heartbeat |

**Entregable**: Cliente Flutter funcional que recibe árbol de widgets y renderiza. 15 widgets base implementados.

---

## Fase D: API Go (Superficie Pública) — 23 semanas

**Archivo**: [07_API_GO.md]

**Depende de**: Fase B (Core Engine funcional), Fase C (widget registry definido)

| Semana | Tarea |
|--------|-------|
| 1-2 | Paquete `fugo` (App, Context, ciclo de vida) |
| 3-5 | Widgets de estructura (Container, Row, Column, Stack, ...) |
| 6-7 | Widgets de contenido (Text, Image, Icon, Divider) |
| 8-10 | Widgets de entrada (Button, TextField, Checkbox, Switch, Slider) |
| 11-12 | Listas y scroll (ListView, GridView) |
| 13-14 | Router y navegación |
| 15-16 | Widgets animados (AnimatedContainer, AnimatedOpacity, AnimatedPositioned) |
| 17-18 | Paquete `fugo/style` (Style, Font, Color, Border, Shadow) |
| 19 | Sistema de eventos |
| 20 | Sistema de Keys |
| 21 | Window Controller |
| 22-23 | Documentación (godoc) + ejemplos |

**Entregable**: API completa, godoc, ejemplos Counter y Router funcionales.

---

## Fase E: Desktop Integration — 15.5 semanas

**Archivo**: [08_DESKTOP.md]

**Depende de**: Fase C (cliente Flutter funcional), Fase D (API Go definida)

| Semana | Tarea |
|--------|-------|
| 1-2 | Window Manager integración (window_manager) |
| 3 | WindowDragArea widget |
| 4-5 | Process Supervisor (os/exec, spawn Flutter desde Go) |
| 6-7 | Signal handling (Linux/macOS/Windows) |
| 8 | Heartbeat / Health check |
| 9-10 | Empaquetado Linux (binario único con Flutter embebido) |
| 11-12 | Empaquetado macOS (.app bundle) |
| 13-14 | Empaquetado Windows (.exe) |
| 15 | File Picker + Clipboard |
| 0.5 | Tests cross-platform |

**Entregable**: `fugo build` produce ejecutable en las 3 plataformas.

---

## Fase F: CLI Tooling — 14.5 semanas

**Archivo**: [06_CLI.md]

**Depende de**: Fase D (API Go), Fase E (empaquetado)

| Semana | Tarea |
|--------|-------|
| 1 | Estructura CLI (Cobra, comandos base) |
| 2-3 | `fugo init` (templates, scaffolding) |
| 4-5 | `fugo run` (compilar + ejecutar, conectar Go↔Flutter) |
| 6-8 | Hot Reload (`--watch`, fsnotify, reinicio automático) |
| 9-11 | `fugo build` (compilación, embed, empaquetado) |
| 12 | `fugo doctor` (verificación de entorno) |
| 13 | `fugo version` + autocompletado |
| 14-14.5 | Tests de integración + documentación de CLI |

**Entregable**: CLI completa con todos los comandos funcionales.

---

## Fase G: Rendimiento — 9.5 semanas

**Archivo**: [09_RENDIMIENTO.md]

**Depende de**: Fase B, C (componentes a optimizar), Fase F (CLI para ejecutar benchmarks)

| Semana | Tarea |
|--------|-------|
| 1-2 | Benchmarks Go (differ, codec, gRPC) |
| 3 | Benchmarks Dart (decode, build) |
| 4 | GC tuning (GOGC, GOMEMLIMIT) |
| 5 | Object pools (VNode, vtprotobuf) |
| 6 | Scheduler avanzado (prioridades de actualización) |
| 7 | CI performance regression |
| 8-9.5 | Profiling y optimización iterativa |

**Entregable**: Benchmarks alcanzan metas. CI bloquea regresiones >10%.

---

## Resumen de tiempos

| Fase | Semanas | Depende de |
|------|---------|------------|
| A — Transporte | 14 | — |
| B — Core SDK | 14 | — (paralelo con A) |
| C — Flutter Client | 13 | A, B |
| D — API Go | 23 | B, C |
| E — Desktop | 15.5 | C, D |
| F — CLI | 14.5 | D, E |
| G — Rendimiento | 9.5 | B, C, F |
| **Total secuencial** | **103.5** | |
| **Total con paralelismo** | **~60-70** | Asumiendo 2 devs, fases A+B en paralelo |

---

## Camino crítico

```
A (14) ──┐
          ├── C (13) ── D (23) ── E (15.5) ── F (14.5) ── G (9.5)
B (14) ──┘
```

**Camino crítico**: A→C→D→E→F→G = 14+13+23+15.5+14.5+9.5 = **89.5 semanas** (~22 meses con 1 dev)

**Con 2 desarrolladores** (paralelizando A+B y D+E parcialmente): ~**14 meses**

---

## Hitos principales

| Hito | Semana aprox | Verificación |
|------|-------------|-------------|
| **H1: Hello Fugo** | 14 | Go envía "Hello" a Flutter por gRPC+FlatBuffers. Flutter lo muestra. |
| **H2: Diff funcional** | 28 | Botón en Flutter → evento a Go → Go diffea → Flutter actualiza texto. |
| **H3: Counter App** | 41 | App Counter funcional: UI en Go, renderizado en Flutter, estado en Go. |
| **H4: 15 widgets** | 51 | Los 15 widgets base funcionan E2E. |
| **H5: Router** | 55 | Navegación entre páginas funcional. |
| **H6: Hot Reload** | 70 | Cambio en .go → reinicio automático <3s. |
| **H7: Build Linux** | 78 | `fugo build` produce binary autocontenido para Linux. |
| **H8: Cross-platform** | 89 | Build funcional en Linux, macOS, Windows. |
| **H9: 60fps garantizado** | 96 | Benchmarks confirman 60fps con 1000 widgets. |
| **H10: v0.2.0 release** | 103.5 | Release completo con SDK, CLI, y documentación. |

---

## Notas sobre el cronograma

1. **Una semana = 40 horas de trabajo concentrado.** No incluye reuniones, documentación adicional, ni investigación exploratoria (esa ya está hecha en [10_INVESTIGACION.md]).

2. **Las fases A y B pueden ejecutarse en paralelo** porque tienen pocas dependencias entre sí. Reducen ~14 semanas del total.

3. **La fase D (API Go) es la más larga** porque implica diseñar, implementar y documentar ~20 widgets con sus builders correspondientes en Flutter.

4. **El risk buffer no está incluido.** En la práctica, añadir 20-30% para imprevistos.

5. **El orden es importante.** No se puede hacer CLI sin API. No se puede hacer Desktop sin Cliente Flutter. Las dependencias están justificadas en cada fase.
