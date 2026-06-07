# 12 — Apéndice

## Glosario

| Término | Definición |
|---------|-----------|
| **SDUI** | Server-Driven UI. Arquitectura donde la lógica de UI reside en un servidor/host y el cliente es un renderizador de instrucciones. |
| **UDS** | Unix Domain Socket. Socket local al sistema operativo que permite comunicación entre procesos sin pasar por la pila TCP/IP. |
| **IPC** | Inter-Process Communication. Mecanismos del SO para comunicación entre procesos. |
| **VDOM** | Virtual DOM. Representación en memoria del árbol de UI que se usa para calcular diferencias (diffs) antes de aplicar cambios al renderizado real. |
| **FlatBuffers** | Formato de serialización binaria zero-copy desarrollado por Google. Permite leer datos sin deserializar. |
| **Protobuf** | Protocol Buffers. Formato de serialización binaria con esquema, usado por gRPC. |
| **vtprotobuf** | Implementación alternativa de Protobuf para Go con codegen unrolled, zero-allocation marshal/unmarshal. |
| **gRPC** | Framework RPC de Google que usa HTTP/2 para transporte y Protobuf para serialización. |
| **Widget Registry** | Patrón Factory que mapea tipos de widget (enum) a builders Dart que construyen el widget Flutter correspondiente. |
| **Reconciler** | Componente que coordina el ciclo: estado cambia → VDOM → Diff → Patch → enviar a Flutter. |
| **Scheduler** | Componente que controla el ritmo de envío de actualizaciones (debounce + alineación de frame). |
| **Patch** | Operación atómica de modificación del árbol de UI: CREATE, UPDATE, DELETE, REPLACE, REORDER. |
| **Background Isolate** | En Dart, un hilo de ejecución separado con su propio heap de memoria, usado para tareas que no deben bloquear el UI thread. |
| **Impeller** | Motor de renderizado de nueva generación de Flutter, reemplazo de Skia con mejor rendimiento en dispositivos modernos. |
| **Pdeathsig** | Flag de Linux (`prctl`) que envía una señal a un proceso hijo cuando el padre muere. Previene procesos zombies. |
| **GOGC** | Variable de entorno de Go que controla el porcentaje de crecimiento del heap que dispara el GC. Default 100. |
| **GOMEMLIMIT** | Variable de entorno de Go (1.19+) que establece un límite suave de memoria. El GC se activa proactivamente para mantenerse bajo este límite. |
| **Mark Assist** | Mecanismo del GC de Go donde goroutines que alocan memoria durante la fase de mark del GC son forzadas a ayudar en el marcado. |
| **Throttle** | Limitar la frecuencia de eventos a un máximo fijo (ej: 1 evento cada 50ms). |
| **Debounce** | Agrupar eventos que ocurren en rápida sucesión, emitiendo solo el último después de un período de inactividad. |
| **Hot Reload** | Capacidad de reemplazar código en ejecución sin perder el estado de la aplicación. En Fugo, se refiere al reinicio rápido del proceso Go con preservación de estado UI. |
| **Embed** | En Go, directiva `//go:embed` que incrusta archivos del sistema de archivos en el binario compilado. |

---

## Riesgos conocidos

### Riesgos técnicos

| Riesgo | Probabilidad | Impacto | Mitigación |
|--------|------------|---------|------------|
| **FlatBuffers + gRPC integración compleja** | Media | Alto | Fase A es la primera. Si resulta demasiado complejo, caer a Protobuf puro (sacrificando zero-copy). |
| **Dart FlatBuffers decode inestable** | Baja | Medio | El paquete oficial (`flat_buffers`) es mantenido por Google. Alternativa: Protobuf en ambos lados. |
| **Flutter Desktop no maduro en Linux** | Alta | Medio | Usar `window_manager` que abstrae diferencias. Tener Windows/macOS como plataformas primarias de testeo. |
| **Hot Reload con estado preservado complejo** | Alta | Bajo | El hot reload es conveniencia, no crítico. Si es muy complejo, aceptar pérdida de estado en reinicio. |
| **GC pauses visibles en Go** | Baja | Alto | `GOGC=200` + `GOMEMLIMIT` + pools minimizan riesgo. Si aún ocurre, investigar Go 1.26+ mejoras de GC. |
| **Empaquetado cross-platform (macOS .app)** | Media | Alto | macOS code signing y notarization pueden ser complejos. Documentar pasos manuales como fallback. |
| **Deriva de esquema FlatBuffers** | Baja | Medio | FlatBuffers soporta forward/backward compatibility. Tests de compatibilidad en CI. |

### Riesgos de proyecto

| Riesgo | Probabilidad | Impacto | Mitigación |
|--------|------------|---------|------------|
| **Alcance creciente (feature creep)** | Alta | Alto | El roadmap define exactamente el scope. Cualquier adición requiere justificación y ajuste de cronograma. |
| **Complejidad subestimada** | Alta | Medio | Las estimaciones son de mejor caso sin buffer. Asumir +20-30% en planificación real. |
| **Dependencia de paquetes externos** | Media | Medio | gRPC, FlatBuffers, window_manager son mantenidos activamente. Tener forks locales si es necesario. |
| **Desarrollador único** | Alta | Alto | El cronograma asume 1 dev. Con 2 se reduce ~40%. Documentar todo para facilitar onboarding. |

---

## Deuda técnica anticipada

Elementos que conscientemente se dejan para después de la release inicial (v0.2.0):

| Elemento | Razón para posponer |
|----------|-------------------|
| **Más de 15 widgets base** | 15 cubren el 90% de casos. Widgets adicionales se añaden por demanda. |
| **Temas y design systems predefinidos** | Fugo es unopinionated en diseño visual. Los temas los construye el desarrollador. |
| **Accesibilidad (a11y)** | Complejo y específico por plataforma. Requiere investigación dedicada. |
| **Internacionalización (i18n)** | Se puede construir sobre paquetes Go existentes. No bloqueante. |
| **Soporte para móviles (Android/iOS)** | Alcance inicial es desktop exclusivamente. |
| **Plugins de terceros** | Fugo no es una plataforma de plugins. Extensibilidad via paquetes Go estándar. |
| **Multi-ventana** | Complejo, requiere manejo de múltiples procesos Flutter. Single-window es suficiente para v0.2.0. |
| **Drag & Drop entre aplicaciones** | Depende de APIs nativas por plataforma. Baja prioridad. |
| **Testing framework integrado** | Se puede testear con el tooling estándar de Go (`go test`). |
| **Editor visual / GUI builder** | Fugo se opera desde terminal. No se planea GUI builder. |

---

## Convenciones del proyecto

### Versionado

Fugo sigue [Semantic Versioning](https://semver.org/):
- **MAJOR** (X.0.0): Cambios incompatibles en la API.
- **MINOR** (0.X.0): Nuevos widgets, funcionalidades backward-compatible.
- **PATCH** (0.0.X): Bug fixes, mejoras de rendimiento.

Versión actual: **0.1.0** (infraestructura). Próximo hito: **0.2.0** (release funcional completa).

### Estilo de código

- **Go**: [Effective Go](https://go.dev/doc/effective_go) + `gofumpt` + `golangci-lint` (config en `.golangci.yml`)
- **Dart**: `flutter_lints` + `dart format`
- **Protobuf/FlatBuffers**: Esquemas en `transport/proto/`

### Commits

- Formato: [Conventional Commits](https://www.conventionalcommits.org/)
- Se usa `lefthook` para pre-commit hooks (golangci-lint, go vet, gofumpt)

---

## Referencias cruzadas dentro del roadmap

| Desde | Hacia | Propósito |
|-------|-------|-----------|
| 00_README.md | Todos | Índice de navegación |
| 01_VISION.md | 02, 10 | Fundamentos y referencias |
| 02_ARQUITECTURA.md | 03, 04, 05, 07, 08 | Arquitectura → implementación |
| 03_CORE_SDK.md | 05, 09, 10 | Core → transporte, rendimiento, investigación |
| 04_FLUTTER_CLIENT.md | 02, 05, 08, 09 | Flutter → arquitectura, transporte, desktop, rendimiento |
| 05_TRANSPORTE.md | 03, 04, 10 | Transporte → core, flutter, investigación |
| 06_CLI.md | 03, 04, 08 | CLI → core, flutter, desktop |
| 07_API_GO.md | 03, 04 | API → core, flutter |
| 08_DESKTOP.md | 04, 06 | Desktop → flutter, CLI |
| 09_RENDIMIENTO.md | 03, 04, 05 | Rendimiento → core, flutter, transporte |
| 10_INVESTIGACION.md | Todos | Compendio de referencias para todas las decisiones |
| 11_CRONOGRAMA.md | 03-09 | Timeline de cada fase |
| 12_APENDICE.md | Todos | Glosario y riesgos de todos los módulos |

---

## Archivos relacionados fuera del roadmap

- `../SPEC.md` — Especificación original del proyecto
- `../docs/ANEXUS_1.md` — Technical Deep Dive & Architectural Relations
- `../docs/ANEXUS_FLUTTER.md` — Flutter Engine Scope & API Mapping
- `../docs/DX_IDEA.md` — Developer Experience & Component Model
- `../docs/DX_OPINIONATED.md` — The Go-Driven Paradigm
- `../CHANGELOG.md` — Historial de cambios
- `../VERSION` — Versión actual (0.1.0)
- `../Makefile` — Comandos de build/test/lint/release
