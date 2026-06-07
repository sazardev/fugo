# 04 — Cliente Flutter

## Alcance

El cliente Flutter es una aplicación Dart precompilada que actúa como **terminal de renderizado** para Fugo. No contiene lógica de negocio, no gestiona estado de aplicación, y no toma decisiones de UI. Su única responsabilidad es:

1. Recibir descripciones de widgets desde Go (vía gRPC + FlatBuffers)
2. Construir el árbol de widgets de Flutter correspondiente
3. Renderizar a 60/120fps usando Impeller
4. Capturar eventos de usuario y enviarlos a Go

[Ver 02_ARQUITECTURA.md:§4 para el contexto arquitectónico completo]

---

## 1. Arquitectura interna del cliente Flutter

```
┌────────────────────────────────────────────┐
│              main.dart (entry point)        │
│  ┌──────────────────────────────────────┐  │
│  │ FugoClientApp (StatelessWidget)       │  │
│  │   └── FugoRenderer (StatefulWidget)   │  │
│  └──────────────────────────────────────┘  │
│                                             │
│  ┌──────────────────────────────────────┐  │
│  │ Background Isolate                     │  │
│  │  ┌──────────────────────────────────┐ │  │
│  │  │ gRPC Client                       │ │  │
│  │  │  ├── Stream<RenderPayload> recv   │ │  │
│  │  │  └── Stream<ClientEvent> send     │ │  │
│  │  └──────────────┬───────────────────┘ │  │
│  │                 │                      │  │
│  │  ┌──────────────▼───────────────────┐ │  │
│  │  │ FlatBuffer Decoder                │ │  │
│  │  │  └── VNode tree → WidgetDescription│ │  │
│  │  └──────────────┬───────────────────┘ │  │
│  └─────────────────┼─────────────────────┘  │
│                    │ SendPort                │
│  ┌─────────────────▼─────────────────────┐  │
│  │ Main Isolate (UI Thread)               │  │
│  │  ┌──────────────────────────────────┐ │  │
│  │  │ Widget Registry                    │ │  │
│  │  │  └── Map<Type, FugoWidgetBuilder> │ │  │
│  │  └──────────────┬───────────────────┘ │  │
│  │                 │                      │  │
│  │  ┌──────────────▼───────────────────┐ │  │
│  │  │ Widget Tree Builder                │ │  │
│  │  │  └── Recursive widget construction│ │  │
│  │  └──────────────┬───────────────────┘ │  │
│  │                 │                      │  │
│  │  ┌──────────────▼───────────────────┐ │  │
│  │  │ Event Debouncer                    │ │  │
│  │  │  └── Throttle + RAF alignment     │ │  │
│  │  └──────────────┬───────────────────┘ │  │
│  │                 │                      │  │
│  │  ┌──────────────▼───────────────────┐ │  │
│  │  │ FugoRenderer (StatefulWidget)      │ │  │
│  │  │  └── build() → Widget tree        │ │  │
│  │  └──────────────────────────────────┘ │  │
│  └───────────────────────────────────────┘  │
└────────────────────────────────────────────┘
```

---

## 2. Background Isolate — Procesamiento de gRPC

### Por qué un isolate separado

Dart es single-threaded por isolate. Si la deserialización de FlatBuffers y la recepción de gRPC ocurrieran en el main isolate (UI thread), cada mensaje competiría con el renderizado por tiempo de CPU, causando jank.

**Solución**: Un isolate de fondo dedicado exclusivamente a:
- Mantener la conexión gRPC bidireccional
- Deserializar FlatBuffers → `WidgetDescription`
- Serializar eventos → FlatBuffer
- Enviar descripciones al main isolate vía `SendPort`

**Referencia**: <https://docs.flutter.dev/perf/isolates> — "You should consider using isolates if your application has computations that are large enough to cause UI jank."

### Implementación

```dart
// main.dart
void main() {
  WidgetsFlutterBinding.ensureInitialized();

  final receivePort = ReceivePort();
  final fugoRenderer = FugoRenderer(receivePort);

  runApp(FugoClientApp(renderer: fugoRenderer));

  // Spawn background isolate for gRPC
  Isolate.spawn(_grpcIsolate, receivePort.sendPort);
}

void _grpcIsolate(SendPort mainSendPort) {
  final client = FugoGrpcClient(mainSendPort);
  client.connect(); // Blocking in this isolate, doesn't affect UI
}

class FugoRenderer extends StatefulWidget {
  final ReceivePort receivePort;

  @override
  State<FugoRenderer> createState() => _FugoRendererState();
}

class _FugoRendererState extends State<FugoRenderer> {
  WidgetRegistry registry = WidgetRegistry();
  WidgetDescription? currentTree;

  @override
  void initState() {
    super.initState();
    widget.receivePort.listen((message) {
      setState(() {
        currentTree = message as WidgetDescription;
      });
    });
  }

  @override
  Widget build(BuildContext context) {
    if (currentTree == null) return Container(); // Loading
    return registry.build(currentTree!);
  }
}
```

---

## 3. Widget Registry — Factory Pattern

### Diseño

El Registry mapea tipos de widget (identificados por un enum `uint32` en FlatBuffer) a builders Dart que saben cómo construir el widget de Flutter correspondiente.

```dart
abstract class FugoWidgetBuilder {
  Widget build(WidgetDescription desc, List<Widget> children);
}

class WidgetRegistry {
  final Map<int, FugoWidgetBuilder> _builders = {};

  void register(int typeId, FugoWidgetBuilder builder) {
    _builders[typeId] = builder;
  }

  Widget build(WidgetDescription desc) {
    final builder = _builders[desc.type];
    if (builder == null) {
      return const SizedBox.shrink(); // Unknown widget type — skip
    }

    final children = desc.children
        .map((childDesc) => build(childDesc))
        .toList();

    return builder.build(desc, children);
  }
}
```

### Widgets base registrados

| Type ID | Widget | Builder |
|---------|--------|---------|
| 0 | Container | `FugoContainerBuilder` |
| 1 | Text | `FugoTextBuilder` |
| 2 | Button | `FugoButtonBuilder` |
| 3 | Row | `FugoRowBuilder` |
| 4 | Column | `FugoColumnBuilder` |
| 5 | Stack | `FugoStackBuilder` |
| 6 | Positioned | `FugoPositionedBuilder` |
| 7 | Expanded | `FugoExpandedBuilder` |
| 8 | Padding | `FugoPaddingBuilder` |
| 9 | Image | `FugoImageBuilder` |
| 10 | TextField | `FugoTextFieldBuilder` |
| 11 | Checkbox | `FugoCheckboxBuilder` |
| 12 | Slider | `FugoSliderBuilder` |
| 13 | ListView | `FugoListViewBuilder` |
| 14 | AnimatedContainer | `FugoAnimatedContainerBuilder` |

### Ejemplo de builder

```dart
class FugoTextBuilder implements FugoWidgetBuilder {
  @override
  Widget build(WidgetDescription desc, List<Widget> children) {
    final props = desc.props; // FlatBuffer bytes

    return Text(
      props.value ?? '',
      style: TextStyle(
        fontSize: props.fontSize?.toDouble(),
        fontWeight: fontWeights[props.fontWeight],
        color: hexToColor(props.color),
        fontFamily: props.fontFamily,
        letterSpacing: props.letterSpacing?.toDouble(),
        height: props.lineHeight?.toDouble(),
      ),
      textAlign: textAligns[props.textAlign],
      overflow: overflows[props.overflow],
    );
  }
}
```

---

## 4. FlatBuffer Decoder — Zero-Copy en Dart

### Esquema FlatBuffer para WidgetDescription

```fbs
namespace fugo.v1;

table WidgetDescription {
  id: uint32;
  type: uint32;
  key: string;
  props: [ubyte];          // Propiedades del widget (varía por tipo)
  children: [uint32];      // IDs de hijos (el builder los resuelve del mapa global)
}
```

### Decodificación zero-copy en Dart

```dart
import 'package:flat_buffers/flat_buffers.dart' as fb;

class WidgetDescription {
  final int id;
  final int type;
  final String? key;
  final fb.BufferContext props;
  final List<int> children;

  WidgetDescription._({
    required this.id,
    required this.type,
    this.key,
    required this.props,
    required this.children,
  });
}
```

La clave es que `props` no se deserializa completamente — se mantiene como `fb.BufferContext` (referencia al buffer original). Solo se extraen campos individuales cuando el builder los necesita. Esto ahorra tiempo de deserialización y memoria.

---

## 5. Widget Tree Builder — Construcción recursiva

### Aplicación de parches (Patch Protocol)

El cliente Flutter recibe parches del Core SDK Go y los aplica incrementalmente:

```dart
class FugoRenderer extends StatefulWidget { ... }

class _FugoRendererState extends State<FugoRenderer> {
  final Map<int, WidgetDescription> _widgetMap = {}; // ID → Description
  int? _rootId;

  void applyPatches(List<Patch> patches) {
    for (final patch in patches) {
      switch (patch.op) {
        case PatchOp.create:
          _widgetMap[patch.node.id] = patch.node;
          if (patch.parentId == 0) _rootId = patch.node.id;
          break;
        case PatchOp.update:
          _widgetMap[patch.nodeId]!.props = patch.props;
          break;
        case PatchOp.delete:
          _deleteRecursive(patch.nodeId);
          break;
        case PatchOp.replace:
          _deleteRecursive(patch.nodeId);
          _widgetMap[patch.node.id] = patch.node;
          break;
        case PatchOp.reorder:
          _widgetMap[patch.nodeId]!.children = patch.children;
          break;
      }
    }
    setState(() {}); // Trigger rebuild con el árbol actualizado
  }

  void _deleteRecursive(int nodeId) {
    final node = _widgetMap.remove(nodeId);
    if (node != null) {
      for (final childId in node.children) {
        _deleteRecursive(childId);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_rootId == null) return const SizedBox.shrink();
    return _buildNode(_rootId!);
  }

  Widget _buildNode(int id) {
    final desc = _widgetMap[id]!;
    final children = desc.children
        .map((childId) => _buildNode(childId))
        .toList();
    return registry.build(desc, children);
  }
}
```

### Optimización con `const` widgets

Flutter permite marcar widgets como `const` para que el framework los reutilice sin reconstruirlos. En Fugo, los widgets cuyas propiedades no cambiaron entre frames se mantienen como la misma instancia, y Flutter cortocircuita el rebuild de ese subárbol.

[Ver 09_RENDIMIENTO.md:§2 para análisis de costo de rebuild]

---

## 6. Event Debouncer — Control de eventos de alta frecuencia

### Problema

Eventos como `onHover` (mouse move) o `onChange` (texto) pueden dispararse cientos de veces por segundo. Enviar cada evento individual a Go saturaría el canal gRPC y desperdiciaría CPU en ambos lados.

### Estrategia de dos capas

```
Evento raw (Flutter)
  │
  ▼
Capa 1: Throttle por tipo
  - onHover: máximo 1 evento cada 50ms (20Hz)
  - onChange (texto): esperar 300ms de inactividad (trailing debounce)
  - onDrag: máximo 1 evento cada 16ms (60Hz, alineado a frame)
  - onClick, onDoubleClick: enviar inmediatamente (eventos discretos)
  │
  ▼
Capa 2: Batch por frame
  - Acumular eventos durante el frame actual (~16ms)
  - Enviar lote en un solo mensaje FlatBuffer al inicio del próximo frame
```

### Implementación

```dart
class EventDebouncer {
  final SendPort _sendPort;
  final Map<String, Timer> _timers = {};
  final List<ClientEvent> _batch = [];

  EventDebouncer(this._sendPort) {
    // Alinear batch al frame de Flutter
    SchedulerBinding.instance.addPostFrameCallback(_flushBatch);
  }

  void onEvent(ClientEvent event) {
    switch (event.type) {
      case 'onHover':
        _throttle(event, Duration(milliseconds: 50));
        break;
      case 'onChange':
        _debounce(event, Duration(milliseconds: 300));
        break;
      case 'onDrag':
        _throttle(event, Duration(milliseconds: 16));
        break;
      default:
        _sendImmediate(event);
    }
  }

  void _throttle(ClientEvent event, Duration interval) {
    if (_timers.containsKey(event.type)) return; // Ya hay uno pendiente
    _sendImmediate(event);
    _timers[event.type] = Timer(interval, () => _timers.remove(event.type));
  }

  void _debounce(ClientEvent event, Duration delay) {
    _timers[event.type]?.cancel();
    _timers[event.type] = Timer(delay, () {
      _timers.remove(event.type);
      _sendImmediate(event);
    });
  }

  void _sendImmediate(ClientEvent event) {
    _batch.add(event);
  }

  void _flushBatch(Duration _) {
    if (_batch.isNotEmpty) {
      _sendPort.send(_batch.toList());
      _batch.clear();
    }
    SchedulerBinding.instance.addPostFrameCallback(_flushBatch);
  }
}
```

---

## 7. Conexión, Reconexión y Heartbeat

### Establecimiento inicial

1. Go host inicia, crea UDS socket en `/tmp/fugo.{pid}.sock`
2. Go host ejecuta `flutter_client` binary con variable de entorno: `FUGO_SOCK=/tmp/fugo.{pid}.sock`
3. Flutter client lee `FUGO_SOCK`, crea canal gRPC sobre ese UDS
4. Handshake gRPC: `Health/Check` → `SERVING`
5. Flutter client inicia stream `RenderStream`
6. Go host envía `FULL_TREE` inicial → Flutter renderiza primera UI

### Heartbeat

Uso del protocolo estándar de health checking de gRPC:

```dart
// Cada 1 segundo
final healthClient = HealthClient(channel);
final response = await healthClient.check(HealthCheckRequest());
if (response.status != HealthCheckResponse_ServingStatus.SERVING) {
  _reconnect();
}
```

[Ver 08_DESKTOP.md:§3 para ciclo de vida completo y manejo de señales]

### Reconexión automática

Si la conexión UDS se pierde (Go process reinicia por hot-reload o crash):

```dart
void _reconnect() async {
  while (true) {
    try {
      await _connect();
      _restoreState(); // Solicitar estado actual a Go
      break;
    } catch (e) {
      await Future.delayed(Duration(milliseconds: 500));
    }
  }
}

void _restoreState() async {
  // Enviar evento especial "Reconnected" a Go
  // Go responde con FULL_TREE del estado actual
  _eventDebouncer.onEvent(ClientEvent(
    type: '_fugo.reconnect',
    data: null,
  ));
}
```

---

## 8. pubspec.yaml y dependencias

```yaml
name: fugo_flutter_client
description: Fugo Flutter rendering client
version: 0.1.0
publish_to: none

environment:
  sdk: '>=3.5.0 <4.0.0'
  flutter: '>=3.24.0'

dependencies:
  flutter:
    sdk: flutter
  grpc: ^4.0.0          # gRPC client
  protobuf: ^3.0.0      # Protobuf runtime (para gRPC service stubs)
  flat_buffers: ^24.0.0  # FlatBuffers runtime
  window_manager: ^0.4.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^4.0.0
```

**Nota sobre `grpc` vs `dart:grpc`**: El paquete `grpc` de Dart proporciona el cliente gRPC. En Flutter, el cliente debe inicializarse en un isolate separado (no en el main isolate que tiene acceso a `dart:ui`).

---

## 9. Estimación de esfuerzo

| Componente | Complejidad | Tiempo estimado |
|-----------|------------|----------------|
| Estructura del proyecto Flutter + pubspec | Baja | 1 semana |
| Background Isolate + gRPC Client | Alta | 2 semanas |
| FlatBuffer Decoder (Dart side) | Media | 2 semanas |
| Widget Registry + Builders base (15 widgets) | Alta | 4 semanas |
| Widget Tree Builder + Patch Application | Media | 2 semanas |
| Event Debouncer | Media | 1 semana |
| Reconnection + Heartbeat | Media | 1 semana |
| **Total Flutter Client** | — | **13 semanas** |

---

## 10. Entregables verificables

- [ ] `flutter_client/` compila y ejecuta en Linux, macOS, Windows
- [ ] 15 builders registrados y funcionales
- [ ] FlatBuffer → Widget tree: cero crashes con árboles de hasta 1000 nodos
- [ ] Event debouncer: mouse move no genera más de 20 eventos/segundo hacia Go
- [ ] Reconexión automática: matar Go process → Flutter espera y reconecta
- [ ] Heartbeat: desconexión detectada en <3 segundos
- [ ] Benchmark: 60fps sostenidos con árbol de 500 widgets recibiendo diffs

---

## Referencias

- Flutter Isolates: <https://docs.flutter.dev/perf/isolates>
- Dart concurrency: <https://dart.dev/language/concurrency>
- gRPC Dart: <https://pub.dev/packages/grpc>
- FlatBuffers Dart: <https://pub.dev/packages/flat_buffers>
- window_manager: <https://pub.dev/packages/window_manager>
- Flutter performance best practices: <https://docs.flutter.dev/perf/best-practices>
- Flutter architectural overview (three trees): <https://docs.flutter.dev/resources/architectural-overview>
