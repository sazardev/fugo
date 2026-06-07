# 03 — Core SDK: Motor Go

## Alcance

El Core SDK es el corazón de Fugo ejecutándose en el proceso Go. Contiene el Virtual DOM, el algoritmo de diffing, el reconciliador que coordina las actualizaciones, y el scheduler que controla el ritmo de envío al cliente Flutter.

**No incluye** la API pública (`fugo/ui`, `fugo/style`) ni la capa de transporte — esas son fases separadas.

[Ver 02_ARQUITECTURA.md para el contexto arquitectónico]

---

## 1. Virtual DOM — Representación plana

### Por qué array plano en lugar de árbol de punteros

**Problema**: React y la mayoría de VDOMs usan árboles de objetos con punteros a hijos (`vnode.children = []*VNode`). Esto causa:
- Mala localidad de caché (cada nodo está en una ubicación de heap distinta)
- Múltiples indirecciones de puntero por traversal
- GC pressure por muchas asignaciones pequeñas

**Solución Fugo**: Array plano (`[]VNode`) donde cada nodo referencia a sus hijos por índice numérico, no por puntero.

```go
type VNode struct {
    ID       uint32   // Identificador estable (no cambia entre frames si el widget es el mismo)
    Parent   uint32   // Índice del nodo padre (0 = root, no tiene padre)
    Type     uint32   // Enum de tipo de widget (Text=1, Container=2, Button=3, ...)
    Key      string   // Key opcional para reconciliación de listas (estilo React key)
    Props    []byte   // FlatBuffer serializado de propiedades (zero-copy comparable)
    Children []uint32 // Índices de nodos hijos en el array
}

type VDOM struct {
    Nodes []VNode       // Array contiguo de todos los nodos
    Root  uint32        // Índice del nodo raíz
    Index map[uint32]int // ID → posición en Nodes para lookup O(1)
}
```

**Ventajas**:
- Todos los nodos en un solo bloque de memoria contiguo → cache-friendly
- Búsqueda por ID en O(1) vía map auxiliar
- Recorrido secuencial sin perseguir punteros
- FlatBuffers serializa arrays de structs naturalmente → zero-copy entre Go y Dart

### Estrategia de IDs estables

Cada widget en la API de Go recibe un ID único y estable:

- **Widgets explícitos**: el desarrollador puede asignar un `Key("myButton")` para IDs deterministas.
- **Widgets anónimos**: el framework genera IDs incrementales (`uint32`). Entre frames, el ID se mantiene si el widget está en la misma posición estructural.
- **Listas**: los items requieren Key para reconciliación eficiente (evita re-crear widgets al reordenar).

[Ver 07_API_GO.md:§5 para el sistema de Keys]

---

## 2. Algoritmo de Diffing — O(n) por ID

### Algoritmo

```
ENTRADA: oldVDOM (árbol anterior), newVDOM (árbol recién generado)
SALIDA:  []Patch (lista de operaciones a aplicar en Flutter)

1. INDEXAR oldVDOM:
   oldMap = map[uint32]*VNode
   for each node in oldVDOM.Nodes:
       oldMap[node.ID] = &node

2. RECORRER newVDOM secuencialmente:
   for each newNode in newVDOM.Nodes:
       oldNode = oldMap[newNode.ID]

       if oldNode == nil:
           EMITIR → Patch{Op: CREATE, Node: newNode}

       else if oldNode.Type != newNode.Type:
           EMITIR → Patch{Op: REPLACE, NodeID: newNode.ID, Node: newNode}

       else:
           if !bytes.Equal(oldNode.Props, newNode.Props):
               EMITIR → Patch{Op: UPDATE, NodeID: newNode.ID, Props: newNode.Props}

           if !slicesEqual(oldNode.Children, newNode.Children):
               EMITIR → Patch{Op: REORDER, NodeID: newNode.ID, Children: newNode.Children}

           delete(oldMap, newNode.ID)  // Marcado como procesado

3. RECOLECTAR ELIMINADOS:
   for each remainingID in oldMap:
       EMITIR → Patch{Op: DELETE, NodeID: remainingID}
```

**Complejidad**:
- Tiempo: O(n) donde n = número de nodos en el nuevo árbol
- Espacio: O(m) donde m = tamaño del mapa de old nodes
- Cada nodo se procesa exactamente una vez

### ¿Por qué O(n) es suficiente?

React usa un algoritmo similar O(n) basado en dos heurísticas:
1. Elementos de diferente tipo producen árboles diferentes
2. Keys estables permiten identidad entre frames

Fugo añade una tercera: **IDs numéricos estables** hacen el lookup O(1) sin necesidad de heurísticas de posición. Esto es más predecible que el diff de React porque no depende del orden de los hijos para identificar nodos.

### Comparación de Props: `bytes.Equal`

Las propiedades de cada widget se serializan a FlatBuffer una sola vez durante la construcción del VDOM. En el diff, la comparación es un simple `bytes.Equal(old.Props, new.Props)` — comparación de slices de bytes, sin deserializar.

Esto es extremadamente rápido (~10-50ns para props de 100-500 bytes) porque:
- No hay deserialización de FlatBuffer durante el diff
- `bytes.Equal` está optimizado en Go con SIMD para slices largos
- Si los slices comparten backing array (misma asignación), devuelve true en O(1)

**Optimización adicional**: Si un widget se construye con los mismos argumentos literales, Go reutiliza el mismo slice subyacente. `bytes.Equal` detecta esto y retorna inmediatamente.

### Caso especial: Listas con keys

Para listas dinámicas (items que se añaden, eliminan o reordenan), Fugo usa keys al estilo React:

```go
for _, item := range items {
    ui.Container(ui.Text(item.Name)).Key(item.ID)
}
```

El algoritmo de diff de listas:
1. Construir `map[key]position` para old children
2. Recorrer new children, buscar key en old map
3. Si la key existe → reordenar (mover), no recrear
4. Si la key no existe → crear
5. Keys sobrantes en old map → eliminar

Esto mantiene el estado de widgets que solo cambiaron de posición (ej: reordenamiento de tabs).

---

## 3. Patch Protocol — Operaciones de diff

```go
type PatchOp uint8

const (
    PatchCreate  PatchOp = 0  // Crear nuevo nodo (con todos sus hijos recursivamente)
    PatchUpdate  PatchOp = 1  // Actualizar propiedades de un nodo existente
    PatchDelete  PatchOp = 2  // Eliminar nodo (y todo su subárbol recursivamente)
    PatchReplace PatchOp = 3  // Reemplazar nodo (tipo diferente, requiere recreación)
    PatchReorder PatchOp = 4  // Reordenar hijos de un nodo
)

type Patch struct {
    Op       PatchOp
    NodeID   uint32
    Node     *VNode      // Solo para Create/Replace
    Props    []byte      // Solo para Update
    Children []uint32    // Solo para Reorder
    ParentID uint32      // Solo para Create (dónde insertar)
    Index    uint32      // Solo para Create (posición entre hermanos)
}
```

### Serialización de Patch

Los patches se serializan como FlatBuffer y se envían por el stream gRPC. Un patch típico ocupa <100 bytes.

Para el render inicial (full tree), se envía un `PatchCreate` único con el nodo raíz y todos sus descendientes embebidos en el mensaje (el cliente Flutter los expande recursivamente).

---

## 4. Reconciler — Coordinador de actualizaciones

El Reconciler es el componente que orquesta el ciclo completo:

```
Estado Go cambia
      │
      ▼
ctx.Update() / ctx.MarkDirty()
      │
      ▼
Re-ejecutar app.Run() closure
      │
      ▼
Nuevo VDOM generado
      │
      ▼
Diffing Engine: oldVDOM vs newVDOM → []Patch
      │
      ▼
Scheduler.debounce(16ms)
      │
      ▼
FlatBuffer encode []Patch
      │
      ▼
gRPC stream.Send(RenderPayload)
```

### ¿Por qué un Scheduler con debounce?

Sin scheduler, cada `ctx.Update()` generaría un envío inmediato a Flutter. Si el desarrollador llama `ctx.Update()` 3 veces en un mismo frame (tres cambios de estado casi simultáneos), se enviarían 3 mensajes separados con 3 diffs parciales.

Con scheduler (debounce de 16ms, alineado al frame):
1. Primer `ctx.Update()` marca dirty flag y programa un tick
2. Segundo y tercer `ctx.Update()` solo actualizan el estado, no re-programan
3. Al cumplirse el tick (~16ms, alineado al próximo frame), se ejecuta UNA sola diffeo con el estado final
4. Se envía UN solo mensaje con el diff neto

Esto evita trabajo redundante y mantiene el ritmo a 60fps.

### Implementación del Scheduler

```go
type Scheduler struct {
    mu       sync.Mutex
    dirty    bool
    timer    *time.Timer
    reconciler *Reconciler
}

func (s *Scheduler) MarkDirty() {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.dirty {
        return  // Ya hay un tick programado
    }
    s.dirty = true
    s.timer.Reset(16 * time.Millisecond)  // Alineado a ~60fps
}

func (s *Scheduler) loop() {
    for range s.timer.C {
        s.mu.Lock()
        s.dirty = false
        s.mu.Unlock()

        s.reconciler.Reconcile()  // Generar VDOM → Diff → Enviar
    }
}
```

[Ver 09_RENDIMIENTO.md:§3 para tuning fino del scheduler]

---

## 5. gRPC Server — Streaming bidireccional

### Definición del servicio (.proto)

```protobuf
syntax = "proto3";
package fugo.v1;

service FugoRender {
  // Canal bidireccional principal
  rpc RenderStream(stream ClientEvent) returns (stream RenderPayload);
}

message RenderPayload {
  PayloadType type = 1;
  repeated Patch patches = 2;
  uint64 seq_num = 3;  // Número de secuencia para ordenamiento
}

message ClientEvent {
  string node_id = 1;
  string event_type = 2;
  bytes event_data = 3;
  uint64 timestamp = 4;
}

message Patch {
  PatchOp op = 1;
  uint32 node_id = 2;
  bytes node_data = 3;    // FlatBuffer serializado del VNode (cuando aplica)
  bytes props_data = 4;   // FlatBuffer de propiedades (UPDATE)
  repeated uint32 children = 5;  // REORDER
  uint32 parent_id = 6;   // CREATE
  uint32 index = 7;       // CREATE
}
```

### Implementación Go del servidor

```go
type RenderServer struct {
    fugo_v1.UnimplementedFugoRenderServer
    reconciler *Reconciler
}

func (s *RenderServer) RenderStream(stream fugo_v1.FugoRender_RenderStreamServer) error {
    // Goroutine 1: Recibir eventos de Flutter y aplicarlos al estado Go
    go func() {
        for {
            event, err := stream.Recv()
            if err != nil {
                return
            }
            s.reconciler.HandleEvent(event)
        }
    }()

    // Goroutine 2 (main): Esperar diffs del Reconciler y enviarlos a Flutter
    for patch := range s.reconciler.PatchChan() {
        payload := &fugo_v1.RenderPayload{
            Type:    determineType(patch),
            Patches: patch,
            SeqNum:  atomic.AddUint64(&seqNum, 1),
        }
        if err := stream.Send(payload); err != nil {
            return err
        }
    }
    return nil
}
```

### vtprotobuf para rendimiento

Para el servidor gRPC en Go, se usa `vtprotobuf` en lugar del codec protobuf estándar:

```go
import (
    "github.com/planetscale/vtprotobuf/codec/grpc"
    "google.golang.org/grpc/encoding"
    _ "google.golang.org/grpc/encoding/proto"
)

func init() {
    encoding.RegisterCodec(grpc.Codec{})
}
```

**Benchmarks (mensajes de ~500B)**:
- `google.golang.org/protobuf`: ~200-400ns marshal, ~48B alloc
- `vtprotobuf`: ~30-80ns marshal, **0 alloc**

[Ver 05_TRANSPORTE.md:§4 para benchmarks completos]

---

## 6. Virtual DOM → FlatBuffer codec

Cada VNode se serializa a FlatBuffer para transmisión. El schema FlatBuffer:

```fbs
namespace fugo.v1;

table VNode {
  id: uint32;
  parent: uint32;
  type: uint32;
  key: string;
  props: [ubyte];       // Propiedades específicas del tipo de widget
  children: [uint32];   // Índices de hijos
}

table VDOM {
  root: uint32;         // Índice del nodo raíz
  nodes: [VNode];       // Array contiguo de nodos
}
```

En Go, la construcción es zero-allocation usando `flatbuffers.Builder`:

```go
func (vd *VDOM) ToFlatBuffer() []byte {
    builder := flatbuffers.NewBuilder(0)  // 0 = grow as needed
    // ... construir el buffer
    return builder.FinishedBytes()
}
```

[Ver 05_TRANSPORTE.md:§3 para especificación completa del codec FlatBuffers]

---

## 7. Estimación de esfuerzo

| Componente | Complejidad | Tiempo estimado |
|-----------|------------|----------------|
| Estructura VDOM + VNode | Baja | 1 semana |
| Algoritmo de Diffing | Alta | 3 semanas |
| Patch Protocol + serialización | Media | 2 semanas |
| Reconciler | Media | 2 semanas |
| Scheduler | Media | 1 semana |
| gRPC Server Definition (.proto) | Baja | 1 semana |
| gRPC Server Go (vtprotobuf) | Media | 2 semanas |
| FlatBuffers Codec (Go side) | Alta | 2 semanas |
| **Total Core SDK** | — | **14 semanas** |

---

## 8. Entregables verificables

- [ ] `engine/vdom.go`: Estructura VNode + VDOM con tests de construcción
- [ ] `engine/differ.go`: Algoritmo de diffing con tests para CREATE, UPDATE, DELETE, REPLACE, REORDER
- [ ] `engine/reconciler.go`: Ciclo completo VDOM → Diff → Patch con tests E2E
- [ ] `engine/scheduler.go`: Debounce de 16ms con tests de múltiples `MarkDirty()`
- [ ] `transport/proto/fugo.proto`: Definición del servicio gRPC
- [ ] `transport/server.go`: Implementación del servidor gRPC con vtprotobuf codec
- [ ] `transport/codec.go`: FlatBuffer marshal/unmarshal para VNode/VDOM
- [ ] Benchmark: diff de árbol de 1000 nodos en <100µs
- [ ] Benchmark: diff de árbol de 100 nodos con 1 cambio en <10µs
- [ ] Test: 1000 updates/segundo sostenidos sin memory leak

---

## Referencias

- React Reconciliation: <https://react.dev/reference/react>
- Snabbdom VDOM (200 SLOC): <https://github.com/snabbdom/snabbdom>
- vtprotobuf: <https://github.com/planetscale/vtprotobuf>
- FlatBuffers Go: <https://github.com/google/flatbuffers>
- FlatBuffers schema guide: <https://flatbuffers.dev/flatbuffers_guide_writing_schema.html>
- gRPC Go bidirectional streaming: <https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc>
- Go GC guide: <https://tip.golang.org/doc/gc-guide>
- bytes.Equal optimization: <https://pkg.go.dev/bytes#Equal>
