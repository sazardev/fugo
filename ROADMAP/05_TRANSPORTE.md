# 05 — Capa de Transporte: IPC, Serialización y Protocolo

## Alcance

Define cómo se comunican el proceso Go y el proceso Flutter: el medio de transporte (Unix Domain Sockets), el formato de serialización (FlatBuffers), la capa RPC (gRPC), y el protocolo de mensajes de aplicación.

[Ver 02_ARQUITECTURA.md para el contexto general]
[Ver 03_CORE_SDK.md para el emisor Go]
[Ver 04_FLUTTER_CLIENT.md para el receptor Dart]

---

## 1. Medio de transporte: Unix Domain Sockets (UDS)

### Decisión: UDS como transporte primario, TCP localhost como fallback

**Plataformas**:
- **Linux**: UDS (`SOCK_STREAM`) — nativo, máxima velocidad
- **macOS**: UDS — soportado, comportamiento idéntico a Linux
- **Windows**: TCP `localhost` — Windows no soporta UDS de forma robusta hasta versiones recientes. TCP localhost es el fallback.

### Benchmarks de latencia (Linux 6.x, CPU moderna)

| Transporte | Round-trip 1 byte | Throughput | Notas |
|-----------|-------------------|------------|-------|
| **UDS** | 5-10µs | 5-10 GB/s | Sin stack TCP |
| **TCP localhost** | 15-25µs | 3-5 GB/s | Stack TCP completo |
| **Named Pipe (FIFO)** | 3-8µs | 1-3 GB/s | Half-duplex |
| **Shared Memory** | 0.1-1µs | 30-50 GB/s | Sin kernel, requiere sincronización |

**Justificación de UDS sobre Shared Memory**:
- Shared memory es ~5-10x más rápido en latencia pura, pero requiere sincronización manual (semáforos, mutexes, ring buffers).
- UDS ya es suficientemente rápido: 10µs de latencia en un presupuesto de frame de 16,667µs = 0.06% del budget.
- gRPC funciona nativamente sobre UDS sin modificación.
- La complejidad de shared memory no se justifica para el ahorro de 9µs.

### Configuración de UDS en Go

```go
import (
    "net"
    "google.golang.org/grpc"
)

func NewServer(socketPath string) (*grpc.Server, net.Listener, error) {
    // Limpiar socket anterior si existe
    os.Remove(socketPath)

    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        return nil, nil, fmt.Errorf("uds listen: %w", err)
    }

    // Permisos restrictivos: solo el usuario actual
    os.Chmod(socketPath, 0600)

    server := grpc.NewServer(
        grpc.KeepaliveParams(keepalive.ServerParameters{
            Time:    10 * time.Second,
            Timeout: 3 * time.Second,
        }),
    )

    return server, listener, nil
}
```

### Configuración en Dart (cliente)

```dart
final channel = ClientChannel(
  socketPath, // '/tmp/fugo.12345.sock'
  options: ChannelOptions(
    credentials: ChannelCredentials.insecure(), // Local, sin TLS
  ),
);
```

---

## 2. Capa RPC: gRPC Bidireccional

### Por qué gRPC en lugar de WebSockets, raw sockets o JSON-RPC

| Alternativa | Pros | Contras | Veredicto |
|------------|------|--------|-----------|
| gRPC | Streaming nativo, health checking, codegen tipado, HTTP/2 | Overhead HTTP/2 (~9B/frame), requiere .proto | ✅ Elegido |
| WebSockets + JSON | Simple, humano-legible | Sin tipado, overhead JSON, sin streaming nativo | ❌ (Flet usa esto — lento) |
| Raw UDS + MessagePack | Sin overhead HTTP/2 | Sin health checking, reinventar rueda de framing | ❌ |
| Raw UDS + FlatBuffers directo | Máxima velocidad | Sin RPC semantics, reinventar routing/errores | ⚠️ Posible optimización futura |

**gRPC se eligió por**:
- Streaming bidireccional nativo (no hay que reinventar multiplexión)
- Health checking estándar (`grpc.health.v1.Health`)
- Codegen tipado en Go y Dart
- Keepalive y reconexión out-of-the-box
- El overhead HTTP/2 es mínimo (9 bytes por frame) y no compite con el presupuesto de 16.67ms

### Servicio gRPC

```protobuf
syntax = "proto3";
package fugo.v1;
option go_package = "github.com/sazardev/fugo/transport/proto/fugo/v1;fugov1";

service FugoRender {
  rpc RenderStream(stream ClientEvent) returns (stream RenderPayload);
}

message RenderPayload {
  PayloadType type = 1;
  repeated Patch patches = 2;
  uint64 seq_num = 3;
}

enum PayloadType {
  FULL_TREE = 0;
  PATCH = 1;
}

message Patch {
  PatchOp op = 1;
  uint32 node_id = 2;
  bytes node_data = 3;
  bytes props_data = 4;
  repeated uint32 children = 5;
  uint32 parent_id = 6;
  uint32 index = 7;
}

enum PatchOp {
  CREATE = 0;
  UPDATE = 1;
  DELETE = 2;
  REPLACE = 3;
  REORDER = 4;
}

message ClientEvent {
  string node_id = 1;
  string event_type = 2;
  bytes event_data = 3;
  uint64 timestamp = 4;
}
```

### Rendimiento de gRPC sobre UDS

**Mensajes pequeños (100-500 bytes)** — caso típico de Fugo:
- Throughput: ~150K-300K mensajes/segundo por core
- Latencia round-trip: ~30-80µs (incluyendo serialización + framing + dispatch)
- A 60fps (16.67ms), se pueden enviar cientos de mensajes por frame

**Mensaje de árbol completo (10KB)**:
- Latencia round-trip: ~100-200µs
- No compite con el presupuesto de frame para el renderizado

**Fuente**: gRPC Go benchmark dashboard — <https://grafana-dot-grpc-testing.appspot.com/>

---

## 3. Serialización de datos: FlatBuffers

### Decisión: FlatBuffers sobre Protobuf para el árbol de widgets

| Característica | Protobuf | FlatBuffers | Impacto en Fugo |
|---------------|----------|-------------|-----------------|
| **Zero-copy read** | ❌ (allocates) | ✅ (pointer arithmetic) | Dart no aloca al leer widgets |
| **Zero-copy write** | ❌ | ⚠️ (builder allocates once) | Go construye y envía sin copia intermedia |
| **Esquema requerido** | ✅ | ✅ | Ambos requieren schema |
| **Dart support** | ✅ (oficial) | ✅ (oficial) | Ambos funcionan |
| **Go support** | ✅ (oficial) | ✅ (oficial) | Ambos funcionan |
| **Evolución de esquema** | ✅ (forward/backward compat) | ✅ (similar) | Ambos permiten añadir campos |
| **Tamaño en wire** | Pequeño (varint) | Similar (vtable + offsets) | Comparable |

**¿Por qué FlatBuffers gana para Fugo?**

El caso de uso crítico es: Go envía un árbol de widgets → Dart lo lee y construye widgets Flutter. Con Protobuf, esto implicaría:
1. Dart recibe bytes
2. Dart deserializa bytes → objetos Dart en el heap (allocations)
3. Dart construye widgets desde esos objetos
4. GC eventualmente recolecta los objetos intermedios

Con FlatBuffers:
1. Dart recibe bytes
2. Dart lee campos directamente del buffer (sin allocations)
3. Dart construye widgets desde los campos leídos
4. No hay objetos intermedios que recolectar

**Esto elimina ~30-50% del tiempo de procesamiento por frame en el lado Dart** y reduce la presión del GC significativamente.

### Esquema FlatBuffers para VDOM

```fbs
namespace fugo.v1;

// Árbol completo (para render inicial)
table VDOM {
  root: uint32;
  nodes: [VNode];
}

// Nodo individual (para actualizaciones incrementales)  
table VNode {
  id: uint32;
  type: uint32;       // WidgetType enum
  parent: uint32;
  key: string;
  props: [ubyte];     // Propiedades específicas del tipo (sub-tabla opaca)
  children: [uint32];
}

// Props para Text widget
table TextProps {
  value: string;
  font_size: float;
  font_weight: uint;
  color: string;
  font_family: string;
  letter_spacing: float;
  line_height: float;
  text_align: uint;
  overflow: uint;
}

// Props para Container widget
table ContainerProps {
  padding_left: float;
  padding_top: float;
  padding_right: float;
  padding_bottom: float;
  margin_left: float;
  margin_top: float;
  margin_right: float;
  margin_bottom: float;
  bg_color: string;
  border_radius: float;
  border_width: float;
  border_color: string;
  width: float;
  height: float;
  alignment: uint;
}

// ... props para cada tipo de widget
```

### Estrategia de serialización: vtprotobuf para gRPC, FlatBuffers para payload

**Doble codec**:
- **Capa gRPC** (headers, health, handshake): `vtprotobuf` — rápido, zero-alloc, nativo del ecosistema gRPC
- **Capa de datos** (árbol de widgets): `FlatBuffers` — zero-copy, embedido en el campo `bytes` del mensaje gRPC

```go
// El mensaje gRPC usa protobuf (vtprotobuf)
type RenderPayload struct {
    Type    PayloadType
    Patches []*Patch
    SeqNum  uint64
}

// Pero el campo Patch.Props contiene FlatBuffer bytes
// y el campo Patch.NodeData contiene un VNode FlatBuffer completo
```

Esto da lo mejor de ambos mundos: la infraestructura de gRPC (conexión, streaming, health) usa Protobuf, pero los datos voluminosos (árboles de widgets) usan FlatBuffers incrustado en campos `bytes`.

---

## 4. Comparativa de serializadores Go (datos de benchmarks)

Fuente: [alecthomas/go_serialization_benchmarks](https://github.com/alecthomas/go_serialization_benchmarks)

| Serializador | Encode (estructura pequeña) | Decode (estructura pequeña) | Allocs por op |
|-------------|---------------------------|---------------------------|---------------|
| **vtprotobuf** | 30-80 ns | 20-60 ns | 0 |
| **golang/protobuf** | 200-400 ns | 150-300 ns | 1-3 |
| **FlatBuffers (Go)** | 100-500 ns* | 10-30 ns | Builder: 1-2, Reader: 0 |
| **MessagePack** | 150-300 ns | 100-200 ns | 1-2 |
| **JSON (stdlib)** | 800-1500 ns | 600-1200 ns | 5-15 |

*FlatBuffers encode incluye construcción del builder. El decode es zero-copy (solo pointer arithmetic).

**Conclusión**: FlatBuffers es significativamente más rápido en decode (el path crítico en Dart) y no aloca en lectura. vtprotobuf es excelente para mensajes de control pequeños.

---

## 5. Protocolo de mensajes de aplicación

### Flujo de mensajes Go → Flutter

```
INITIAL RENDER:
  RenderPayload {
    type: FULL_TREE,
    patches: [
      Patch { op: CREATE, node_id: 0, node_data: <VNode FlatBuffer completo> }
    ],
    seq_num: 0
  }

SUBSEQUENT UPDATES:
  RenderPayload {
    type: PATCH,
    patches: [
      Patch { op: UPDATE, node_id: 42, props_data: <TextProps FlatBuffer> },
      Patch { op: DELETE, node_id: 55 },
    ],
    seq_num: 1
  }
```

### Flujo de mensajes Flutter → Go

```
USER CLICK:
  ClientEvent {
    node_id: "button_7",
    event_type: "onClick",
    event_data: <flatbuffer con timestamp y metadata>,
    timestamp: 1717803123456
  }

TEXT INPUT (debounced):
  ClientEvent {
    node_id: "textfield_3",
    event_type: "onChange",
    event_data: <flatbuffer con {value: "Hola"} >,
    timestamp: 1717803123890
  }

WINDOW RESIZE (throttled):
  ClientEvent {
    node_id: "_window",
    event_type: "onResize",
    event_data: <flatbuffer con {width: 1024, height: 768}>,
    timestamp: 1717803124234
  }
```

### Números de secuencia y ordenamiento

Cada `RenderPayload` lleva un `seq_num` monótonamente creciente. Esto permite al cliente Flutter:
- Detectar mensajes perdidos (gap en seq_num → solicitar FULL_TREE)
- Ignorar mensajes duplicados o fuera de orden
- Sincronizar estado en reconexión ("mi último seq_num fue 42, envíame desde ahí")

---

## 6. Estimación de esfuerzo

| Componente | Complejidad | Tiempo estimado |
|-----------|------------|----------------|
| Esquema FlatBuffer (.fbs completo) | Media | 2 semanas |
| Codec FlatBuffer Go (marshal/unmarshal) | Alta | 3 semanas |
| Codec FlatBuffer Dart (decode zero-copy) | Media | 2 semanas |
| Definición .proto + codegen | Baja | 1 semana |
| gRPC Server Go (con vtprotobuf) | Media | 2 semanas |
| gRPC Client Dart | Media | 1 semana |
| UDS setup cross-platform (Linux/macOS) | Media | 1 semana |
| TCP fallback para Windows | Media | 1 semana |
| Health checking + keepalive | Baja | 1 semana |
| **Total Transporte** | — | **14 semanas** |

---

## 7. Entregables verificables

- [ ] `.fbs` completo con todos los tipos de widget y sus propiedades
- [ ] FlatBuffer marshal en Go: zero-allocation para árboles de hasta 1000 nodos
- [ ] FlatBuffer decode en Dart: zero-copy reads verificados con benchmark
- [ ] gRPC RenderStream funcional: Go→Flutter y Flutter→Go
- [ ] UDS: Go y Flutter se comunican sobre socket Unix
- [ ] TCP fallback: misma comunicación sobre localhost TCP en Windows
- [ ] Health check: cliente detecta servidor caído en <3s
- [ ] Reconnection: secuencia completa de reconexión automática
- [ ] Benchmark: latencia round-trip Go↔Flutter <200µs en mensajes de <1KB
- [ ] Benchmark: throughput >100K mensajes/segundo en streaming bidireccional

---

## Referencias

- gRPC concepts: <https://grpc.io/docs/what-is-grpc/core-concepts/>
- gRPC Go: <https://grpc.io/docs/languages/go/>
- gRPC Dart: <https://pub.dev/packages/grpc>
- gRPC health checking: <https://grpc.io/docs/guides/health-checking/>
- FlatBuffers: <https://flatbuffers.dev/>
- FlatBuffers Go: <https://github.com/google/flatbuffers>
- FlatBuffers Dart: <https://pub.dev/packages/flat_buffers>
- vtprotobuf: <https://github.com/planetscale/vtprotobuf>
- go_serialization_benchmarks: <https://github.com/alecthomas/go_serialization_benchmarks>
- Unix domain sockets: Linux `unix(7)` man page
- gRPC benchmarking: <https://grpc.io/docs/guides/benchmarking/>
