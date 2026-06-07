# 10 — Compendio de Investigación

## Propósito

Este documento consolida toda la investigación realizada para fundamentar las decisiones técnicas de Fugo. Cada sección incluye referencias verificables, alternativas consideradas y descartadas, y lecciones aprendidas de proyectos similares.

No es un documento de "benchmarks bonitos" — es un registro de ingeniería con datos reales y fuentes rastreables.

[Este documento es referenciado por todos los demás archivos del roadmap]

---

## 1. SDUI: Lo que funciona y lo que no

### Proyectos estudiados

| Proyecto | Stack | Modelo SDUI | Lección para Fugo |
|----------|-------|------------|-------------------|
| **Flet** | Python + Flutter | JSON/WebSocket | SDUI con Flutter es viable. JSON es el cuello de botella. GIL de Python mata concurrencia. |
| **Phoenix LiveView** | Elixir + HTML | Diff HTML/WebSocket | Modelo de "estado en servidor + diffs" es correcto. Latencia de red es aceptable para web, inaceptable para sensación "nativa". |
| **Airbnb Epoxy** | Android/Kotlin | Modelos tipados + hash diff | El patrón de modelos tipados con ID estable es el enfoque correcto para SDUI. |
| **Shopify SDUI** | React Native + JSON | JSON full screens | SDUI remoto falla por latencia y payloads grandes. JSON parsing en móvil es lento. |
| **LinkedIn Voyager** | Native mobile | JSON schema-driven | Explosión de complejidad: demasiados tipos de widget. Miles de tipos = inmantenible. |
| **Etsy Entropy** | Mobile | JSON UI blobs | Sin caché posible con UI completamente dinámica. Necesario permitir análisis estático. |

### Conclusiones para Fugo

1. **SDUI local (IPC) elimina el problema de latencia.** La razón #1 por la que SDUI remoto falla es la red. UDS local = 5-10µs.
2. **Mantener la superficie de widgets pequeña.** ~15 tipos de widgets primitivos, componer el resto. No caer en la trampa de Voyager (miles de tipos).
3. **Formato binario, no JSON.** FlatBuffers zero-copy evita el overhead de parsing que mató a Etsy y Shopify.
4. **El modelo LiveView (diff, no full tree) es correcto.** Enviar solo cambios, no el árbol completo en cada frame.
5. **Tipado fuerte en el contrato.** Protobuf/FlatBuffers con esquema previene errores de runtime que JSON no detecta.

### Referencias

- Flet: <https://flet.dev/docs/>, <https://github.com/flet-dev/flet>
- Phoenix LiveView: <https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.html>
- Airbnb Epoxy: <https://github.com/airbnb/epoxy>
- Shopify SDUI: <https://shopify.engineering/>
- LinkedIn Voyager: <https://engineering.linkedin.com/blog/2017/05/building-a-server-driven-ui-framework>

---

## 2. IPC: Comparativa de transportes locales

### Benchmarks de latencia (round-trip)

| Transporte | Latencia | Throughput | Complejidad |
|-----------|----------|------------|-------------|
| Unix Domain Sockets | 5-10µs | 5-10 GB/s | Baja (socket API estándar) |
| TCP localhost | 15-25µs | 3-5 GB/s | Baja |
| Shared Memory (mmap) | 0.1-1µs | 30-50 GB/s | Alta (sincronización manual) |
| Named Pipes (FIFO) | 3-8µs | 1-3 GB/s | Media (half-duplex) |
| Message Queues (POSIX) | 5-10µs | 0.5-1 GB/s | Media |

### ¿Por qué no Shared Memory?

Aunque shm es ~10x más rápido en latencia pura, la diferencia es 9µs — en un presupuesto de frame de 16,667µs, esto es 0.05% del budget. No justifica la complejidad adicional:
- Sincronización manual (semáforos, mutexes, ring buffers)
- Sin soporte nativo en gRPC
- Más frágil (corrupción de memoria, condiciones de carrera)

**Decisión**: UDS con gRPC. Suficientemente rápido y significativamente más simple.

### Referencias

- Linux `unix(7)` man page
- gRPC UDS transport: <https://grpc.io/docs/what-is-grpc/core-concepts/#unix-domain-sockets>

---

## 3. Serialización: FlatBuffers vs Protobuf vs JSON vs MessagePack

### Comparativa detallada

| Característica | JSON | MessagePack | Protobuf | FlatBuffers | Cap'n Proto |
|---------------|------|-------------|----------|-------------|-------------|
| **Zero-copy read** | ❌ | ❌ | ❌ | ✅ | ✅ |
| **Zero-copy write** | ❌ | ❌ | ❌ | ⚠️ | ✅ |
| **Schema** | ❌ | ❌ | ✅ | ✅ | ✅ |
| **Go support** | stdlib | 3rd party | ✅ oficial | ✅ oficial | 3rd party |
| **Dart support** | stdlib | 3rd party | ✅ oficial | ✅ oficial | ❌ |
| **Backward compat** | — | — | ✅ | ✅ | ✅ |
| **Tamaño wire** | Grande | Mediano | Pequeño | Pequeño | Pequeño |

### Benchmarks Go (estructura 150 bytes, de alecthomas/go_serialization_benchmarks)

| Serializador | Marshal | Unmarshal | Allocs/op |
|-------------|---------|-----------|-----------|
| encoding/json | 800-1500ns | 600-1200ns | 5-15 |
| msgpack | 150-300ns | 100-200ns | 1-2 |
| protobuf (google) | 200-400ns | 150-300ns | 1-3 |
| vtprotobuf | 30-80ns | 20-60ns | 0 |
| FlatBuffers | 100-500ns* | 10-30ns | 1-2 / 0** |

*FlatBuffers marshal incluye construcción del builder. **FlatBuffers unmarshal es zero-copy: 0 allocs.

### Por qué FlatBuffers gana para el árbol de widgets

El caso crítico es **Dart leyendo el árbol**. Con Protobuf, cada mensaje se deserializa a objetos Dart → GC pressure. Con FlatBuffers, Dart lee campos directamente del buffer → 0 allocs en el path caliente.

Para mensajes de control (gRPC headers, health checks), Protobuf (vtprotobuf) es más que suficiente y más simple de integrar con gRPC.

### Referencias

- go_serialization_benchmarks: <https://github.com/alecthomas/go_serialization_benchmarks>
- FlatBuffers: <https://flatbuffers.dev/>
- vtprotobuf: <https://github.com/planetscale/vtprotobuf>

---

## 4. Virtual DOM y Diffing

### Algoritmos estudiados

| Algoritmo | Complejidad | Usado por |
|-----------|------------|-----------|
| React Reconciliation | O(n) | React, Preact |
| Snabbdom | O(n) | Vue 2, Snabbdom |
| Inferno | O(n) optimizado | Inferno |
| LitHTML | O(n) template-based | Lit |
| Morphdom | O(n) DOM-based | Phoenix LiveView |

### React Reconciliation (algoritmo base)

React asume dos heurísticas:
1. Elementos de diferente tipo producen árboles diferentes → reemplazo completo.
2. Keys estables permiten identidad entre renders.

Complejidad O(n) con n = número de nodos. En la práctica, React tarda ~0.5-2ms para 10,000 nodos en JavaScript (V8).

### Snabbdom (implementación de referencia)

Snabbdom (~200 líneas de código) demuestra que el VDOM diffing puede ser extremadamente compacto. Introduce:
- **Módulos**: hooks (`create`, `update`, `destroy`) que se disparan en fases del diff.
- **Thunks**: memoización de subárboles. Si los datos de entrada no cambiaron, skip.
- **Keyed children**: reconciliación O(n) de hijos con keys mediante mapa key→índice.

### Fugo VDOM: ¿Por qué una implementación propia?

Los VDOM existentes (React, Snabbdom, VirtualDOM) están diseñados para DOM/HTML. Fugo necesita:
- FlatBuffers como formato de serialización (no HTML strings)
- Array plano de nodos (no árbol de punteros) para localidad de caché
- Props como `[]byte` opacos (no objetos tipados) para diff rápido con `bytes.Equal`
- IDs numéricos estables (no dependientes de posición en el árbol)

Ningún VDOM existente cumple estos requisitos. La implementación propia (~500 LOC) es más simple que adaptar uno existente.

### Referencias

- React Reconciliation: <https://react.dev/reference/react>
- Snabbdom: <https://github.com/snabbdom/snabbdom>
- Inferno: <https://github.com/infernojs/inferno>
- Morphdom: <https://github.com/patrick-steele-idem/morphdom>

---

## 5. Flutter como motor de renderizado headless

### ¿Puede Flutter funcionar como "dumb terminal"?

**Sí**, con tres enfoques posibles:

1. **Platform Channel**: La app Dart recibe instrucciones por platform channel, construye widgets. Es el enfoque de Flet y el recomendado para Fugo.

2. **External Textures**: Go escribe pixel data en GPU buffer, Flutter lo muestra como `Texture` widget. Útil para rasterizado pero bypassa el sistema de widgets de Flutter.

3. **Platform Views**: Embebe vistas nativas dentro de Flutter. Requiere implementación por plataforma (GTK, Win32, NSView). Complejo.

**Decisión**: Platform Channel (enfoque 1). Máxima flexibilidad con mínima complejidad. Se mantiene dentro del ecosistema Flutter estándar.

### Flutter Embedder API

La Flutter Engine expone una API C (`flutter_embedder.h`) para embeber Flutter en cualquier aplicación. Pero Fugo NO necesita usar esta API directamente porque Flutter ya corre como aplicación desktop standalone. La integración es a nivel de platform channel Dart↔Go (vía gRPC), no a nivel de embedder C.

### Flutter Three-Tree Architecture

Flutter mantiene tres árboles:
1. **Widget tree**: Configuración inmutable, se recrea cada frame.
2. **Element tree**: Mutable, persistente. Flutter reconcilia widgets viejos vs nuevos.
3. **RenderObject tree**: Layout y pintado.

Cuando Fugo envía un diff, solo se actualizan los Widgets afectados. Flutter minimiza el rebuild usando `runtimeType` + `key` para identificar qué Elementos se actualizan vs recrean. `const` widgets se canonicalizan — misma instancia entre frames.

### Referencias

- Flutter architectural overview: <https://docs.flutter.dev/resources/architectural-overview>
- Flutter embedder API: <https://github.com/flutter/engine/blob/main/shell/platform/embedder/embedder.h>
- flutter-pi (embedder headless): <https://github.com/ardera/flutter-pi>
- Flet architecture: <https://flet.dev/docs/>

---

## 6. gRPC Performance

### gRPC Go streaming benchmarks

Datos de gRPC performance dashboard (<https://grafana-dot-grpc-testing.appspot.com/>) y benchmarks comunitarios:

| Escenario | Mensajes/segundo | Notas |
|-----------|-----------------|-------|
| Unary RPC (1KB) | ~50K/s | Por core |
| Streaming (1KB) | ~150-300K/s | Por core, bidireccional |
| Streaming (100B) | ~500K-1M/s | Mensajes pequeños |
| Latencia P99 unary (UDS) | ~80µs | 1KB mensaje |
| Latencia P99 stream (UDS) | ~30µs | 1KB mensaje |

**Conclusión**: gRPC sobre UDS puede manejar fácilmente el tráfico de Fugo (máximo ~60 diffs/segundo a 60fps, ~1000 eventos/segundo). El overhead es insignificante comparado con el renderizado.

### vtprotobuf vs protobuf estándar

| Métrica | google.golang.org/protobuf | vtprotobuf |
|---------|---------------------------|------------|
| Marshal (500B) | ~200-400ns, 48B alloc | ~30-80ns, 0 alloc |
| Unmarshal (500B) | ~150-300ns, 96B alloc | ~20-60ns, 0 alloc |
| Size (500B) | ~50-100ns | ~10-20ns |

vtprotobuf genera código unrolled (sin reflection) que es 3-8x más rápido en marshal. Para el path caliente de Fugo (cada frame serializa un diff), esta diferencia es significativa.

### Referencias

- gRPC benchmarking: <https://grpc.io/docs/guides/benchmarking/>
- gRPC Go performance: <https://github.com/grpc/grpc-go/blob/master/Documentation/benchmark.md>
- vtprotobuf: <https://github.com/planetscale/vtprotobuf>

---

## 7. Hot Reload en Go

### Técnicas estudiadas

| Técnica | Plataformas | Tiempo | Complejidad |
|---------|------------|--------|-------------|
| Recompilar y re-ejecutar | Todas | 1-3s | Baja |
| Go plugin (`plugin.Open`) | Solo Linux | <1s | Alta (frágil) |
| hashicorp/go-plugin | Todas | 1-2s | Media |
| Re-ejecución con `syscall.Exec` | Linux, macOS | <500ms | Media |
| Interpretado (yaegi) | Todas | <100ms | Alta (no production) |

### Decisión: Recompilar y re-ejecutar (estilo Air)

Es la opción más simple, funciona en todas las plataformas, y el tiempo de 1-3s es aceptable. El estado de UI se preserva mediante serialización en el lado Flutter.

**Alternativa futura**: `syscall.Exec` para reinicio más rápido (reemplaza el proceso en caliente, mantiene PID y file descriptors). Requiere investigación adicional para Windows (no soporta `Exec`).

### Referencias

- Air: <https://github.com/air-verse/air>
- Go plugin: <https://pkg.go.dev/plugin>
- hashicorp/go-plugin: <https://github.com/hashicorp/go-plugin>
- yaegi: <https://github.com/traefik/yaegi>

---

## 8. GC y memoria en aplicaciones desktop

### Go GC en aplicaciones interactivas

Go 1.19+ tiene un GC concurrente mark-sweep con pausas STW típicamente sub-100µs. Sin embargo:

- **Mark assist**: Goroutines que alocan durante la fase de mark son forzadas a ayudar. Esto añade latencia a operaciones que normalmente serían instantáneas.
- **Solución**: `GOGC=200` reduce frecuencia de GC. `GOMEMLIMIT=256MiB` previene OOM en sistemas con poca RAM.

### Dart GC en Flutter

Dart usa GC generacional:
- **Young space** (Scavenger): Rápido, STW, <1ms típico.
- **Old space**: Concurrente mark-sweep, similar al GC de Go.

La creación masiva de widgets en cada frame va al young space y se recolecta rápido. El riesgo es promover objetos al old space (referencias retenidas).

### Referencias

- Go GC Guide: <https://tip.golang.org/doc/gc-guide>
- Dart VM GC: <https://mrale.ph/dartvm/>
- Dart GC wiki: <https://github.com/dart-lang/sdk/wiki/Garbage-Collection>

---

## 9. Lecciones de proyectos similares

### Flet (Python + Flutter)

- **Acierto**: Demostró que SDUI con Flutter es viable y atractivo para devs backend.
- **Error**: JSON/WebSocket limita el rendimiento. El GIL de Python limita concurrencia.
- **Lección**: Usar formato binario + goroutines elimina ambos problemas.

### Phoenix LiveView

- **Acierto**: El modelo de "diff mínimo sobre conexión persistente" funciona.
- **Error**: Limitado a HTML/CSS, no puede aprovechar renderizado nativo.
- **Lección**: Aplicar el mismo principio de diffs a widgets nativos Flutter.

### Wails (Go + WebView)

- **Acierto**: Excelente DX para Go developers. Bindings Go↔JS automáticos.
- **Error**: WebView como motor de renderizado = sub-rendimiento vs nativo. No hay acceso a widgets nativos.
- **Lección**: Fugo evita WebView completamente. Flutter/Impeller es renderizado nativo.

### Tauri (Rust + WebView)

- **Acierto**: Binarios pequeños, rendimiento superior a Electron.
- **Error**: Sigue siendo WebView con las limitaciones de CSS/DOM para UIs complejas.
- **Lección**: Fugo comparte la filosofía de "backend en lenguaje de sistemas + GUI separada", pero con Flutter en vez de WebView.

### Referencias

- Flet: <https://flet.dev/>
- Wails: <https://wails.io/>
- Tauri: <https://tauri.app/>
- Phoenix LiveView: <https://hexdocs.pm/phoenix_live_view>

---

## Nota final

Esta investigación fundamenta cada decisión en los documentos del roadmap. Cuando un documento dice "UDS sobre TCP" o "FlatBuffers sobre JSON", la justificación completa está aquí. No son opiniones — son decisiones de ingeniería basadas en datos, benchmarks, y lecciones de la industria.
