# 00 — Índice y Navegación del Roadmap

> **⚠️ Nota de estado (léeme primero).** Este roadmap describe el **diseño original**. El código implementado diverge en dos puntos importantes:
> 1. **Serialización:** no se usa FlatBuffers ni vtprotobuf, sino **Protocol Buffers estándar** (`google.golang.org/protobuf`), con las props de cada nodo como mensaje protobuf anidado en un campo `bytes`. No hay lecturas zero-copy ni códecs zero-alloc. Los presupuestos de rendimiento (ver `09`) se **re-basaron sobre mediciones reales** (`engine/perf_test.go`, `engine/alloc_test.go`, `engine/differ_bench_test.go`).
> 2. **API / paquete de widgets:** el paquete vivo es **`fg/`** con constructores **sin prefijo** (`fg.Text(...)`), no `fugo/ui` ni `ui.NewText(...)`.
>
> Cuando este documento y el código discrepen, **manda el código**.

## Propósito

Este directorio contiene el plan de construcción completo para Fugo: un framework de interfaz gráfica para desktop que permite escribir aplicaciones **exclusivamente en Go**, delegando el renderizado a un motor Flutter precompilado mediante comunicación Server-Driven UI (SDUI) por IPC local.

No es un plan de producto ni un MVP incremental. Es una **hoja de ruta técnica** con análisis, fundamentos, referencias y cronograma para construir una herramienta funcional y eficiente.

## Cómo navegar este roadmap

Cada archivo aborda una dimensión del proyecto. Se recomienda leer en orden numérico la primera vez, luego usar como referencia por módulo.

| # | Archivo | Qué contiene | Depende de |
|---|---------|-------------|------------|
| 00 | `00_README.md` | Este índice | — |
| 01 | `01_VISION.md` | Visión general, problema que resuelve, por qué Go+Flutter | — |
| 02 | `02_ARQUITECTURA.md` | Arquitectura completa del sistema, decisiones de diseño | 01 |
| 03 | `03_CORE_SDK.md` | Go SDK: protobuf, Virtual DOM, diffing engine, gRPC server | 02, 10 |
| 04 | `04_FLUTTER_CLIENT.md` | Cliente Flutter: widget registry, event handling, pipeline de renderizado | 02, 05 |
| 05 | `05_TRANSPORTE.md` | Capa de transporte: UDS, FlatBuffers, protocolo de mensajes, serialización | 02, 10 |
| 06 | `06_CLI.md` | Herramienta CLI: `fugo init`, `fugo run`, `fugo build`, hot-reload | 03, 04, 08 |
| 07 | `07_API_GO.md` | API pública de Go: widgets, estilos, layouts, eventos, routing, state | 03, 04 |
| 08 | `08_DESKTOP.md` | Integración desktop: window management, ciclo de vida, empaquetado | 04, 06 |
| 09 | `09_RENDIMIENTO.md` | Rendimiento: benchmarks, GC tuning, estrategias de diffing y debouncing | 03, 04, 05 |
| 10 | `10_INVESTIGACION.md` | Compendio de investigación: referencias, alternativas descartadas, lecciones | — |
| 11 | `11_CRONOGRAMA.md` | Timeline estimado, dependencias entre fases, hitos | 03-09 |
| 12 | `12_APENDICE.md` | Glosario, riesgos conocidos, deuda técnica anticipada | Todos |

## Convenciones

- **Referencias cruzadas**: se usa `[Ver 03_CORE_SDK.md]` para navegar entre documentos.
- **Referencias externas**: cada afirmación técnica incluye enlace a fuente (paper, benchmark, repo, docs oficiales).
- **Decisiones**: se marcan con **Decisión:** y se justifican con alternativas consideradas.
- **Estimaciones**: en semanas de trabajo concentrado (no meses-hombre diluidos).

## Estado actual del proyecto

- **Versión**: 0.2.0 — el motor (diff/reconciler/scheduler con prioridad), la API de widgets
  (`fg/`, 30 tipos), el transporte gRPC (con token de auth opcional), el supervisor, la CLI y el
  cliente Flutter están **implementados y funcionan end-to-end**.
- **Fase E (Desktop)**: control de ventana en runtime (`window_manager`), `WindowDragArea`,
  portapapeles y diálogos de archivo nativos (`Context.Clipboard()` / `Context.Files()`),
  empaquetado en las 3 plataformas. Pendiente: nada bloqueante.
- **Fase G (Rendimiento)**: diff con object-pool, GC tuning (`FUGO_GOGC` / `FUGO_GOMEMLIMIT`),
  benchmarks Go **y** Dart, y un gate de regresión en CI (1ms / 1000 nodos, zero-alloc en el
  fast-path). Pendiente: scheduler con prioridades múltiples más allá de normal/inmediata.
- **Módulo Go**: `github.com/sazardev/fugo` (Go 1.26.3)
- **Documentación base**: `CLAUDE.md` es la guía canónica y actualizada; `SPEC.md` + `docs/` son diseño.

> **El código difiere de este roadmap en dos puntos deliberados:**
> 1. **Serialización**: el roadmap describe **FlatBuffers** (ver `05_TRANSPORTE.md`); la
>    implementación usa **Protocol Buffers estándar** (`google.golang.org/protobuf`), con los
>    props de cada widget marshaled como protobuf anidado dentro de `WidgetNode.props`.
> 2. **API de widgets**: el paquete es `fg/` (no `ui/`) y los constructores no llevan prefijo
>    `New` — `fg.Text(...)`, `fg.Button(...)`, `fg.Container(...)`, devolviendo `*fg.XxxWidget`.

## Filosofía del roadmap

1. **Sin MVP ni versiones incrementales para vender**: se construye completo, fase por fase.
2. **Sin IA ni integraciones especulativas**: el foco es Go + Flutter SDUI.
3. **Decisiones basadas en datos**: benchmarks reales, no opiniones.
4. **Cada fase produce algo ejecutable y verificable**: no hay fases de "paperwork".
5. **El rendimiento es requisito de primer orden**, no optimización tardía.
