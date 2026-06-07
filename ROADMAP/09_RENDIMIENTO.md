# 09 — Rendimiento y Optimización

## Alcance

Define las estrategias de rendimiento que garantizan que Fugo compita con aplicaciones nativas. Incluye benchmarks esperados, tuning de garbage collection, optimizaciones de diffing, estrategias de debouncing, y métricas de calidad objetivo.

**El rendimiento en Fugo no es una optimización tardía — es un requisito de arquitectura desde el día cero.**

[Ver 03_CORE_SDK.md para implementación del VDOM y diffing]
[Ver 05_TRANSPORTE.md para benchmarks de IPC y serialización]
[Ver 04_FLUTTER_CLIENT.md:§6 para debouncing en Flutter]

---

## 1. Budget de frame y metas de rendimiento

### El presupuesto de 60fps

```
Frame budget total: 16,667 µs (16.67 ms)
                                   |
    ┌──────────────────────────────┤
    │                              │
Go side                    Flutter side
├── Diffing: <100µs        ├── Build: <2ms
├── FlatBuffer encode:     ├── Layout: <3ms
│   <50µs                  ├── Paint: <3ms
├── gRPC send: <10µs       ├── GPU raster: <5ms
├── UDS transport: <10µs   └── Total Flutter: <13ms
├── gRPC recv: <15µs
├── FlatBuffer decode: <5µs
└── Total Go→Flutter: <200µs

TOTAL CRÍTICO (Go+transporte): ~200µs = 1.2% del budget
TOTAL FLUTTER: ~8-13ms = 48-78% del budget
MARGEN: ~3-5ms para imprevistos
```

**Meta de rendimiento primaria**: Go nunca debe contribuir más del 2% al tiempo de frame. La latencia Go→Flutter debe ser <200µs en el percentil 99.

### ¿Qué pasa si no se cumple?

Si Go tarda más de 2ms en difiar + serializar, se come el margen de Flutter y la app droppea frames. Esto se previene con:
- VDOM en array plano (O(n) con constantes bajas)
- FlatBuffers zero-copy en ambos lados
- Scheduler con debounce (máximo 1 diff por frame)
- vtprotobuf zero-alloc para mensajes gRPC

---

## 2. Diffing Benchmark Target

### Escenarios de benchmark

| Escenario | Tamaño de árbol | Cambios | Meta de tiempo |
|-----------|----------------|---------|---------------|
| Sin cambios | 1000 nodos | 0 | <10µs (short-circuit) |
| 1 cambio de propiedad | 100 nodos | 1 UPDATE | <10µs |
| 1 cambio de propiedad | 1000 nodos | 1 UPDATE | <50µs |
| 50 cambios | 1000 nodos | 50 UPDATE | <100µs |
| Reordenamiento de lista | 500 nodos | 100 REORDER | <200µs |
| Árbol completamente nuevo | 500 nodos | 500 CREATE | <500µs |
| Render inicial (FULL_TREE) | 1000 nodos | 1000 CREATE | <1ms |

### Short-circuit: cuando no hay cambios

El caso más común: el desarrollador llama `ctx.Update()` pero ningún estado cambió. En este caso:

```go
// Optimización: comparar hash del árbol completo antes de difiar
func (r *Reconciler) shouldDiff(oldVDOM, newVDOM *VDOM) bool {
    if oldVDOM.Hash == newVDOM.Hash {
        return false  // Nada cambió, skip completo
    }
    return true
}
```

El hash del árbol se calcula incrementalmente durante la construcción del VDOM (hash de todos los Props + estructura). Si el hash coincide, se omite el diffing completo — ahorrando ~50µs-1ms.

---

## 3. Scheduler Tuning

### Debounce y alineación de frame

```
ctx.Update() calls:   ──┬─┬──┬─┬─┬──────────────────
                         │ │  │ │ │
Scheduler:               ▼ ▼  ▼ ▼ ▼
                         └─┴──┴─┴─┘ (batched)
                                   │
Frame boundary (16.67ms): ─────────┼────────────────
                                   ▼
                              1 solo diff
```

El scheduler acumula llamadas a `Update()` y ejecuta UNA sola reconciliación por frame (~cada 16ms). Esto es crítico porque:

- Si el desarrollador llama `ctx.Update()` 10 veces en un loop, solo se diffea UNA vez con el estado final.
- Si un evento de mouse causa `onHover` → `ctx.Update()` 60 veces/segundo, el scheduler lo limita a 1 por frame.

### Scheduler avanzado: prioridades

Para el futuro, se pueden implementar prioridades de actualización:

```go
const (
    PriorityImmediate = 0  // Click, keypress (debe responder ya)
    PriorityHigh      = 1  // Input text change
    PriorityNormal    = 2  // Data update
    PriorityLow       = 3  // Background refresh
)

func (c *Context) UpdateWithPriority(prio int) {
    c.scheduler.MarkDirty(prio)
}
```

Las actualizaciones de prioridad `Immediate` fuerzan un diff síncrono (sin esperar el próximo frame), útil para feedback táctil inmediato (botones, toggles).

---

## 4. Estrategia de memoria y GC

### Go GC Tuning

**Problema**: El GC de Go es concurrente, pero durante la fase de mark, goroutines que alocan memoria son forzadas a hacer "mark assist" — trabajo de GC que añade latencia a operaciones como `stream.Send()`.

**Configuración recomendada**:

```bash
# En desarrollo y producción
GOGC=200          # GC se dispara cuando el heap triplica (no duplica)
GOMEMLIMIT=256MiB # Límite suave de memoria total
```

**Efecto**: Menos ciclos de GC (más memoria, más espacio entre colecciones). Para una app desktop típica, 256MiB es insignificante y reduce la frecuencia de GC de ~2/s a ~1/s.

### Pool de objetos VNode

```go
var vnodePool = sync.Pool{
    New: func() interface{} {
        return &VNode{}
    },
}

func NewVNode() *VNode {
    vn := vnodePool.Get().(*VNode)
    vn.reset()
    return vn
}

func (vn *VNode) Release() {
    vnodePool.Put(vn)
}
```

Las pools reducen las asignaciones de heap en la construcción del VDOM. Los nodos se reciclan entre frames en lugar de ser garbage-collected.

### vtprotobuf Pools

```go
import "github.com/planetscale/vtprotobuf/pool"

var renderPayloadPool = pool.NewRenderPayloadPool()

func (s *Server) sendPatch(patch *Patch) {
    payload := renderPayloadPool.Get()
    defer renderPayloadPool.Put(payload)

    payload.Type = fugov1.PayloadType_PATCH
    payload.SeqNum = atomic.AddUint64(&seqNum, 1)
    // ... construir payload ...

    stream.Send(payload)
    // payload se devuelve al pool automáticamente al Put()
}
```

### Dart GC — Minimizar allocs en el UI thread

- **`const` widgets**: Los builders deben usar `const` para widgets cuyas propiedades son literales.
- **FlatBuffer zero-copy**: Las props se mantienen como `fb.BufferContext` (buffer view), no se deserializan a objetos Dart.
- **Pool de WidgetDescription**: Reciclar objetos de descripción de widgets entre frames.

---

## 5. Debouncing y throttling de eventos

### Matriz de estrategias por tipo de evento

| Evento | Frecuencia raw | Estrategia | Frecuencia resultante |
|--------|---------------|-----------|----------------------|
| `onClick` | Esporádico | Inmediato | ~1-10/s |
| `onDoubleClick` | Esporádico | Inmediato | ~1-5/s |
| `onLongPress` | Esporádico | Inmediato | ~1-5/s |
| `onHover` (mouse) | 60-1000/s | Throttle 50ms | 20/s |
| `onDrag` (pan) | 60-120/s | Throttle 16ms | 60/s |
| `onChange` (texto) | 10-60/s | Debounce 300ms | ~3/s |
| `onResize` (ventana) | 10-60/s | Throttle 100ms | 10/s |
| `onScroll` | 60-120/s | Throttle 16ms | 60/s |

**Justificación**:
- Eventos discretos (click, longpress): enviar inmediatamente. El costo de latencia para feedback táctil es inaceptable.
- Eventos continuos de alta frecuencia (hover, drag, scroll): throttle a la tasa mínima necesaria. Para animaciones fluidas, 60Hz es suficiente; para hover (tooltips), 20Hz basta.
- Entrada de texto (onChange): trailing debounce. Esperar a que el usuario deje de escribir por 300ms antes de enviar a Go. Esto evita enviar 20 eventos para una palabra de 5 letras.

[Ver 04_FLUTTER_CLIENT.md:§6 para la implementación del Event Debouncer en Dart]

---

## 6. Optimizaciones del canal gRPC

### Keepalive tuning

```go
// Go server
grpc.KeepaliveParams(keepalive.ServerParameters{
    Time:              10 * time.Second,  // Ping cada 10s
    Timeout:            3 * time.Second,   // Timeout para pong
    MaxConnectionIdle:  30 * time.Second,  // Cerrar si idle 30s
    MaxConnectionAge:   0,                 // Sin límite
    MaxConnectionAgeGrace: 0,
})
```

```go
// Go server enforcement policy
grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
    MinTime:             5 * time.Second,  // Mínimo entre pings del cliente
    PermitWithoutStream: true,             // Permitir health checks sin stream
})
```

### Tamaño de mensaje

gRPC por defecto limita mensajes a 4MB. Para Fugo, esto es más que suficiente:

- Mensaje de diff típico: 50-500 bytes
- FULL_TREE de 1000 nodos: ~50-100KB
- FULL_TREE de 5000 nodos: ~250-500KB

Se puede aumentar el límite a 8MB para seguridad:

```go
grpc.MaxRecvMsgSize(8 * 1024 * 1024),
grpc.MaxSendMsgSize(8 * 1024 * 1024),
```

### Compresión

Para mensajes de diff (pequeños), la compresión no se justifica (overhead > ahorro). Para FULL_TREE (grandes, ~100KB+), se puede habilitar gzip:

```go
grpc.UseCompressor(gzip.Name),
```

Pero en la práctica, sobre UDS local, la compresión añade latencia de CPU sin beneficiar el throughput (UDS ya es más rápido que cualquier compresor). **Decisión**: sin compresión en el canal local.

---

## 7. Benchmark suite

### Benchmarks Go

```go
// engine/differ_bench_test.go
func BenchmarkDiffNoChanges(b *testing.B) {
    old := buildVDOM(1000)    // Árbol de 1000 nodos
    new := buildVDOM(1000)    // Mismo árbol
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Diff(old, new)
    }
}

func BenchmarkDiffOneChange(b *testing.B) { ... }
func BenchmarkDiffFullReplace(b *testing.B) { ... }
func BenchmarkFlatBufferEncode(b *testing.B) { ... }
func BenchmarkGRPCRoundTrip(b *testing.B) { ... }
```

### Benchmarks Dart

```dart
// test/benchmark/flatbuffer_decode_bench_test.dart
void main() {
  test('FlatBuffer decode 1000 nodes', () {
    final bytes = generateFlatBufferTree(1000);
    benchmark('decode', () {
      decodeWidgetTree(bytes);
    });
  });
}
```

### CI Performance Regression

En CI, ejecutar benchmarks y comparar contra línea base:

```yaml
# .github/workflows/bench.yml
- name: Run benchmarks
  run: go test -bench=. -benchmem ./engine/... | tee bench.txt
- name: Compare with baseline
  run: benchstat baseline.txt bench.txt
```

Si algún benchmark degrada >10%, la PR se bloquea.

---

## 8. Perfil de memoria esperado

### Go process

| Componente | Memoria estimada |
|-----------|-----------------|
| Go runtime + GC metadata | ~5-10 MB |
| VDOM (1000 nodos) | ~100-200 KB |
| gRPC server + buffers | ~2-5 MB |
| FlatBuffer builder | ~1-2 MB |
| Pool de objetos | ~1 MB |
| **Total Go (~1000 widgets)** | **~15-25 MB** |

### Flutter process

| Componente | Memoria estimada |
|-----------|-----------------|
| Dart VM + runtime | ~20-30 MB |
| Flutter Engine (Impeller) | ~30-50 MB |
| Widget tree (1000 widgets) | ~5-10 MB |
| GPU textures/buffers | ~20-40 MB |
| **Total Flutter** | **~80-130 MB** |

**Total combinado**: ~100-150 MB para una app desktop con ~1000 widgets. Comparable a una app Electron pequeña, pero con renderizado nativo y sin Chromium.

---

## 9. Estimación de esfuerzo

| Componente | Complejidad | Tiempo estimado |
|-----------|------------|----------------|
| Benchmarks Go (differ, codec, gRPC) | Media | 2 semanas |
| Benchmarks Dart (decode, build) | Media | 1 semana |
| GC tuning (GOGC, GOMEMLIMIT) | Baja | 0.5 semanas |
| Object pools (VNode, vtprotobuf) | Media | 1 semana |
| Scheduler avanzado (prioridades) | Media | 1 semana |
| CI performance regression | Media | 1 semana |
| Profiling y optimización iterativa | Alta | 3 semanas |
| **Total Rendimiento** | — | **9.5 semanas** |

---

## 10. Entregables verificables

- [ ] Diff de 1000 nodos sin cambios: <10µs
- [ ] Diff de 1000 nodos con 50 cambios: <100µs
- [ ] FlatBuffer encode 1000 nodos: <500µs
- [ ] FlatBuffer decode 1000 nodos (Dart): <200µs
- [ ] gRPC round-trip (1KB): <200µs P99
- [ ] 60fps sostenidos con árbol de 1000 widgets y 10 cambios/frame
- [ ] Sin memory leaks en test de 1 hora (1000 updates/segundo)
- [ ] Go heap <30MB estable tras 30 minutos de uso
- [ ] CI bloquea PRs que degradan benchmarks >10%

---

## Referencias

- Go GC Guide: <https://tip.golang.org/doc/gc-guide>
- Go sync.Pool: <https://pkg.go.dev/sync#Pool>
- vtprotobuf pool: <https://github.com/planetscale/vtprotobuf>
- Flutter performance best practices: <https://docs.flutter.dev/perf/best-practices>
- Flutter rendering performance: <https://docs.flutter.dev/perf/rendering-performance>
- gRPC keepalive: <https://grpc.io/docs/guides/keepalive/>
- benchstat: <https://pkg.go.dev/golang.org/x/perf/cmd/benchstat>
- Google Docs OT algorithm: <https://en.wikipedia.org/wiki/Operational_transformation>
- Figma multiplayer tech: <https://www.figma.com/blog/how-figmas-multiplayer-technology-works/>
