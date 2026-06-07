# 06 — CLI Tooling

## Alcance

La herramienta de línea de comandos `fugo` es el centro de operaciones del desarrollador. Todo el ciclo de desarrollo — crear, ejecutar, compilar, empaquetar — se realiza desde la terminal, sin asistentes gráficos ni configuraciones fuera del código Go.

[Ver SPEC.md:§4 para la especificación base del CLI]

---

## 1. Comandos

### `fugo init <name>`

Crea la estructura base del proyecto:

```
myapp/
├── main.go
├── go.mod           (module myapp, require github.com/sazardev/fugo)
├── go.sum
├── .gitignore
└── README.md        (template mínimo)
```

**`main.go` generado**:

```go
package main

import (
    "github.com/sazardev/fugo"
    "github.com/sazardev/fugo/ui"
)

func main() {
    app := fugo.NewApp(fugo.AppOptions{
        Title:  "myapp",
        Width:  800,
        Height: 600,
    })

    app.Run(func(ctx *fugo.Context) ui.Widget {
        return ui.Container(
            ui.Center(
                ui.Text("Hello, Fugo!").FontSize(24),
            ),
        ).Fill()
    })
}
```

**Flags**:
- `--module-path <path>`: Ruta del módulo Go (default: inferido del nombre)

### `fugo run`

Inicia el servidor Go, levanta el proceso Flutter, y conecta ambos.

**Flujo**:
1. Compilar Go binary (modo dev: sin optimizaciones para compilación rápida)
2. Encontrar UDS libre: `/tmp/fugo.{pid}.sock`
3. Iniciar gRPC server Go en el socket
4. Ejecutar `fugo_flutter` binary con `FUGO_SOCK=/tmp/fugo.{pid}.sock`
5. Esperar handshake gRPC
6. Enviar árbol inicial FULL_TREE
7. Entrar en loop de eventos

**Flags**:
- `--watch`: Activar hot-reload (watchea archivos .go)
- `--port <n>`: Puerto TCP para Windows (en vez de UDS)
- `--verbose`: Mostrar logs de conexión, diffs, eventos

### `fugo build`

Compila la aplicación en un entregable autocontenido.

**Flujo**:
1. Compilar Go binary con `-ldflags="-s -w"` (stripped, sin debug info)
2. Compilar Flutter client en modo Release (AOT)
3. Embeber Flutter binary como recurso en Go (usando `embed.FS`)
4. Generar ejecutable final

**Output**:
- Linux: `build/myapp` (binary único con Flutter embebido)
- macOS: `build/myapp.app` (bundle)
- Windows: `build/myapp.exe` (con `fugo_flutter.exe` al lado o embebido)

**Flags**:
- `--target <os>`: Cross-compilar para otro OS
- `--output <path>`: Directorio de salida

### `fugo doctor`

Verifica que el entorno de desarrollo esté correctamente configurado:

```
✓ Go 1.26.3
✓ Flutter SDK 3.24+
✓ Dart SDK 3.5+
✓ gRPC tools (protoc)
✓ FlatBuffers compiler (flatc)
✓ Git
```

### `fugo version`

Imprime la versión del CLI y del SDK:

```
fugo CLI v0.1.0 (commit: abc1234, built: 2026-06-07)
fugo SDK v0.1.0
Flutter engine: Impeller 3.24
Go: 1.26.3
```

---

## 2. Hot Reload

### Comportamiento

Con `fugo run --watch`, el CLI monitorea cambios en archivos `.go` del proyecto:

```
Watching /home/user/myapp/...
[15:34:21] main.go changed
[15:34:22] Recompiling...
[15:34:23] Restarting Go server...
[15:34:23] Flutter reconnected
[15:34:23] State restored ✓
```

### Arquitectura del Hot Reload

```
┌────────────────────────────────────────────┐
│              CLI Process (fugo)             │
│  ┌──────────────────┐  ┌────────────────┐  │
│  │ File Watcher      │  │ Process Manager│  │
│  │ (fsnotify)        │  │                │  │
│  │                   │  │ 1. Kill old Go │  │
│  │ onChange: main.go │──▶ 2. go build    │  │
│  │                   │  │ 3. Start new Go│  │
│  └──────────────────┘  └────────────────┘  │
└────────────────────────────────────────────┘
         │
         ▼
┌────────────────────────────────────────────┐
│              Go Process (old)               │
│  1. Recibe SIGTERM                         │
│  2. Envía SaveState a Flutter              │
│  3. Cierra gRPC stream                     │
│  4. Exit(0)                                │
└────────────────────────────────────────────┘
         │
         ▼
┌────────────────────────────────────────────┐
│              Go Process (new)               │
│  1. Crea nuevo UDS socket                  │
│  2. Inicia gRPC server                     │
│  3. Espera conexión Flutter                │
│  4. Recibe RestoreState                    │
│  5. Envía FULL_TREE actualizado            │
└────────────────────────────────────────────┘
         │
         ▼
┌────────────────────────────────────────────┐
│           Flutter Process (persiste)        │
│  1. Detecta cierre de stream gRPC          │
│  2. Guarda estado actual (widget map)      │
│  3. Watch UDS socket file                  │
│  4. Reconecta cuando aparece nuevo socket  │
│  5. Envía RestoreState event               │
│  6. Recibe FULL_TREE → reconstruye UI      │
└────────────────────────────────────────────┘
```

### Tiempos esperados

| Paso | Tiempo |
|------|--------|
| Detectar cambio de archivo | ~10ms (fsnotify) |
| `go build` (proyecto típico) | ~500ms-2s |
| Kill old Go process | ~50ms |
| Start new Go + gRPC server | ~100ms |
| Flutter reconnect | ~50ms |
| State restore + re-render | ~16ms (1 frame) |
| **Total (típico)** | **~1-3 segundos** |

Es más lento que Flutter Hot Reload (~300ms) porque Go requiere recompilación completa (no tiene JIT como Dart VM en modo debug). Pero 1-3 segundos es aceptable para el ciclo de desarrollo.

### Limitaciones

- Solo se aplica a cambios en código Go. Cambios en el cliente Flutter requieren rebuild del cliente.
- El estado de la UI se preserva, pero el estado de memoria Go (variables, conexiones DB) se pierde en el reinicio. Para preservar estado entre reinicios, el desarrollador puede implementar serialización en `app.OnSaveState()` y `app.OnRestoreState()`.

---

## 3. Empaquetado y distribución

### Estrategia de embedding del cliente Flutter

```go
//go:embed flutter_client/build/*
var flutterClient embed.FS
```

En `fugo build`, el binario Go contiene el cliente Flutter precompilado embebido. Al ejecutarse:
1. Go extrae el binario Flutter a un directorio temporal (`/tmp/fugo_{pid}/`)
2. Lo ejecuta como subproceso
3. Al cerrar, limpia los archivos temporales

### Alternativa: side-by-side

En algunos casos (especialmente Windows), puede ser preferible distribuir el cliente Flutter como archivo separado en el mismo directorio. El flag `--embed=false` controla esto.

### Estructura del binario final

```
myapp (Go binary)
├── Código Go compilado (lógica de negocio + Fugo SDK)
├── gRPC server
├── Process supervisor
└── [Embedded: fugo_flutter binary (≈15-20MB)]
```

Tamaño estimado del binario final: **~20-30MB** (Go ~5-10MB + Flutter engine ~15-20MB).

---

## 4. Dependencias del CLI

```go
// go.mod del CLI
require (
    github.com/sazardev/fugo v0.1.0         // SDK principal
    github.com/fsnotify/fsnotify v1.7.0     // File watching
    github.com/fatih/color v1.18.0          // Colored terminal output
    github.com/schollz/progressbar/v3 v3.19.0 // Progress bars (build)
    github.com/spf13/cobra v1.8.0           // CLI framework
    google.golang.org/grpc v1.64.0          // gRPC (para health check, etc.)
)
```

### Framework CLI: Cobra

Se usa [spf13/cobra](https://github.com/spf13/cobra) por ser el estándar de facto en Go y facilitar:
- Subcomandos (`fugo init`, `fugo run`, `fugo build`)
- Flags persistentes y locales
- Autocompletado de shell
- Generación de docs/man pages

---

## 5. Integración con lefthook y CI

El proyecto Fugo ya usa lefthook para git hooks. El comando `fugo` debe integrarse:

```bash
# .lefthook.yml (adición futura)
pre-commit:
  commands:
    fugo-doctor:
      run: fugo doctor --ci
```

En CI, `fugo doctor --ci` verifica el entorno sin salida interactiva y retorna código de salida no cero si algo falta.

---

## 6. Estimación de esfuerzo

| Componente | Complejidad | Tiempo estimado |
|-----------|------------|----------------|
| Estructura CLI (Cobra) | Baja | 1 semana |
| `fugo init` (templates, scaffolding) | Media | 2 semanas |
| `fugo run` (compilar + ejecutar) | Media | 2 semanas |
| Hot Reload (`--watch`) | Alta | 3 semanas |
| `fugo build` (compilación + embed) | Alta | 3 semanas |
| `fugo doctor` | Baja | 1 semana |
| `fugo version` | Trivial | 0.5 semanas |
| Empaquetado cross-platform | Media | 2 semanas |
| **Total CLI** | — | **14.5 semanas** |

---

## 7. Entregables verificables

- [ ] `fugo init myapp` crea proyecto funcional
- [ ] `fugo run` inicia app con ventana Flutter visible
- [ ] `fugo run --watch` detecta cambios y hace hot reload
- [ ] `fugo build` genera binario autocontenido
- [ ] `fugo doctor` reporta estado del entorno
- [ ] `fugo version` imprime versiones correctas
- [ ] Binario final ejecutable en Linux, macOS, Windows (sin dependencias externas)
- [ ] Hot reload: ciclo completo <3s para proyecto de prueba
- [ ] Autocompletado de shell funcional

---

## Referencias

- Cobra CLI framework: <https://github.com/spf13/cobra>
- fsnotify: <https://github.com/fsnotify/fsnotify>
- Air (Go hot reload): <https://github.com/air-verse/air>
- Flutter Hot Reload: <https://docs.flutter.dev/tools/hot-reload>
- Go embed: <https://pkg.go.dev/embed>
- SPEC.md: `../SPEC.md:§4`
