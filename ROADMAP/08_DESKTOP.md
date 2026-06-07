# 08 — Integración Desktop

## Alcance

Cubre todo lo relacionado con la ejecución de Fugo como aplicación de escritorio: gestión de ventanas, ciclo de vida del subproceso Flutter, manejo de señales del sistema operativo, empaquetado en binarios autocontenidos, y distribución cross-platform.

[Ver 02_ARQUITECTURA.md:§5 para el Supervisor de procesos]
[Ver 04_FLUTTER_CLIENT.md:§7 para el lado cliente de la conexión]
[Ver 06_CLI.md:§3 para el empaquetado desde el CLI]

---

## 1. Gestión de ventanas — window_manager

### Por qué window_manager

El ecosistema Flutter Desktop tiene tres opciones principales para control de ventanas:

| Paquete | Linux | Windows | macOS | Frameless | Custom title bar |
|---------|-------|---------|-------|-----------|-----------------|
| `window_manager` (leanflutter) | GTK | Win32 | Cocoa | ✅ | ✅ |
| `bitsdojo_window` | GTK | Win32 | Cocoa | ✅ | ✅ |
| Flutter nativo | Parcial | Parcial | Parcial | ❌ | ❌ |

**Decisión**: `window_manager` ([pub.dev](https://pub.dev/packages/window_manager), score 88.88)

**Justificación**: API unificada cross-platform, mantenimiento activo, soporte para frameless windows con custom chrome via widgets Flutter, menor configuración nativa por plataforma que `bitsdojo_window`.

### Configuración en el cliente Flutter

```dart
import 'package:window_manager/window_manager.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await windowManager.ensureInitialized();

  // Configuración inicial mínima — Go enviará la real luego
  WindowOptions windowOptions = WindowOptions(
    size: Size(800, 600),
    center: true,
    titleBarStyle: TitleBarStyle.hidden,  // Fugo controla el chrome
    skipTaskbar: false,
  );

  windowManager.waitUntilReadyToShow(windowOptions, () async {
    await windowManager.show();
    await windowManager.focus();
  });

  // ... resto de la inicialización Fugo
}
```

### Mapeo WindowController Go → window_manager Dart

| Go API | Dart window_manager |
|--------|-------------------|
| `Window().Minimize()` | `windowManager.minimize()` |
| `Window().Maximize()` | `windowManager.maximize()` |
| `Window().Close()` | `windowManager.close()` |
| `Window().SetTitle(t)` | `windowManager.setTitle(t)` |
| `Window().SetSize(w, h)` | `windowManager.setSize(Size(w, h))` |
| `Window().Center()` | `windowManager.center()` |
| `Window().SetFullScreen(b)` | `windowManager.setFullScreen(b)` |
| `Window().SetFrameless(b)` | `windowManager.setAsFrameless()` |

### WindowDragArea — Arrastre de ventana custom

Cuando se usa `TitleBarStyle.hidden`, Fugo expone un widget para crear barras de título personalizadas:

```go
// En Go:
ui.WindowDragArea(
    ui.Row(
        ui.Text("Mi App").FontSize(14),
        // ... botones de minimizar, maximizar, cerrar
    ),
)
```

En Flutter, esto mapea a:

```dart
GestureDetector(
  onPanStart: (_) => windowManager.startDragging(),
  child: childWidget,
)
```

---

## 2. Ciclo de vida del subproceso

### Inicio del proceso Flutter desde Go

```go
package supervisor

import (
    "context"
    "fmt"
    "os"
    "os/exec"
    "syscall"
)

type FlutterProcess struct {
    cmd     *exec.Cmd
    sockPath string
}

func StartFlutter(ctx context.Context, sockPath, flutterBinary string) (*FlutterProcess, error) {
    cmd := exec.CommandContext(ctx, flutterBinary)

    // Pasar el socket path como variable de entorno
    cmd.Env = append(os.Environ(),
        fmt.Sprintf("FUGO_SOCK=%s", sockPath),
        "FUGO_VERBOSE=0",
    )

    // Redirigir stdout/stderr para debugging
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // Grupo de procesos separado para señalización limpia
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid:   true,                    // Nuevo grupo de procesos
        Pdeathsig: syscall.SIGTERM,         // Linux: kernel mata si padre muere
    }

    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("start flutter: %w", err)
    }

    return &FlutterProcess{
        cmd:     cmd,
        sockPath: sockPath,
    }, nil
}
```

### Señales del sistema operativo

Go debe manejar señales para una terminación limpia:

```go
func (s *FlutterProcess) WaitForSignal(ctx context.Context) error {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

    select {
    case sig := <-sigCh:
        log.Printf("received signal %v, shutting down", sig)
        return s.Shutdown(5 * time.Second)

    case <-ctx.Done():
        return s.Shutdown(5 * time.Second)
    }
}

func (s *FlutterProcess) Shutdown(timeout time.Duration) error {
    // 1. Enviar SIGTERM al grupo de procesos
    syscall.Kill(-s.cmd.Process.Pid, syscall.SIGTERM)

    // 2. Esperar con timeout
    done := make(chan error, 1)
    go func() {
        done <- s.cmd.Wait()
    }()

    select {
    case err := <-done:
        log.Println("flutter process exited cleanly")
        s.cleanup()
        return err

    case <-time.After(timeout):
        log.Println("flutter didn't exit, sending SIGKILL")
        syscall.Kill(-s.cmd.Process.Pid, syscall.SIGKILL)
        <-done
        s.cleanup()
        return fmt.Errorf("flutter process killed after timeout")
    }
}

func (s *FlutterProcess) cleanup() {
    os.Remove(s.sockPath) // Limpiar socket UDS
}
```

### Heartbeat y Health Checking

Se usa el protocolo estándar de gRPC health checking:

```go
import (
    "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)

func setupHealthCheck(server *grpc.Server) {
    healthServer := health.NewServer()
    grpc_health_v1.RegisterHealthServer(server, healthServer)
    healthServer.SetServingStatus("fugo.v1.FugoRender", grpc_health_v1.HealthCheckResponse_SERVING)
}
```

En el lado Dart, el cliente verifica periódicamente:

```dart
final healthClient = HealthClient(channel);
Timer.periodic(Duration(seconds: 1), (_) async {
  try {
    final resp = await healthClient.check(
      HealthCheckRequest()..service = 'fugo.v1.FugoRender',
    );
    if (resp.status != HealthCheckResponse_ServingStatus.SERVING) {
      _handleDisconnection();
    }
  } catch (e) {
    _handleDisconnection();
  }
});
```

### Escenarios de terminación

| Escenario | Comportamiento |
|-----------|---------------|
| Usuario cierra ventana Flutter | Flutter envía `_window.close` event a Go → Go cierra gRPC stream → Go hace shutdown → ambos procesos terminan |
| Usuario hace Ctrl+C en terminal | Go recibe SIGINT → Go envía SIGTERM a Flutter via grupo de procesos → Flutter cierra → Go limpia y termina |
| Go panics | `Pdeathsig: SIGTERM` (Linux) → kernel mata Flutter automáticamente |
| Flutter crash | Canal gRPC se rompe → heartbeat timeout → Go detecta y termina |
| Sistema operativo se apaga | SIGTERM a ambos procesos → shutdown normal |

### Prevención de zombies

- `cmd.Wait()` se llama siempre en una goroutine o vía el `Shutdown`.
- `Pdeathsig` asegura que si Go muere sin limpiar, el kernel mata Flutter.
- En macOS, `Pdeathsig` no existe — se usa `kqueue`/`dispatch_source` para monitorear el proceso padre. Alternativamente, heartbeat con timeout.

```go
// Fallback para macOS/Windows (sin Pdeathsig)
go func() {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    for range ticker.C {
        if parentIsDead() {
            os.Exit(0)
        }
    }
}()
```

---

## 3. Empaquetado cross-platform

### Estrategia de distribución

```
fugo build → produce:
  Linux:   build/myapp                  (binario único con Flutter embebido)
  macOS:   build/myapp.app/             (bundle .app)
  Windows: build/myapp.exe              (+ fugo_flutter.exe al lado)
```

### Embedding del cliente Flutter en Go

```go
package main

import (
    "embed"
    "os"
    "path/filepath"
)

//go:embed flutter_client/build/linux/x64/release/bundle/*
var flutterBundle embed.FS

func extractFlutter() (string, error) {
    tmpDir, err := os.MkdirTemp("", "fugo_flutter_*")
    if err != nil {
        return "", err
    }

    // Copiar archivos del embed al disco
    err = copyEmbedDir(flutterBundle, "flutter_client/build/linux/x64/release/bundle", tmpDir)
    if err != nil {
        return "", err
    }

    binary := filepath.Join(tmpDir, "fugo_flutter")
    os.Chmod(binary, 0755)
    return binary, nil
}
```

### Compilación del cliente Flutter

```bash
# Linux
flutter build linux --release

# macOS  
flutter build macos --release

# Windows
flutter build windows --release
```

El output de cada plataforma se coloca en `flutter_client/build/{os}/release/bundle/`.

### Manejo de assets y fuentes

Las fuentes y assets que el desarrollador Go referencia deben estar disponibles para Flutter:

1. **Fuentes**: El desarrollador las declara en Go (`style.Font("Inter", ...)`). Fugo las registra en el `pubspec.yaml` del cliente Flutter durante `fugo build` o las carga dinámicamente en runtime.
2. **Imágenes**: `ui.ImageAsset("logo.png")` → Go envía los bytes de la imagen por gRPC (como base64 en FlatBuffer) o los copia al bundle de Flutter durante el build.

---

## 4. Diálogos del sistema

### File Picker

```go
// Go API
path, err := ctx.ShowOpenDialog(fugo.FileDialogOptions{
    Title: "Seleccionar archivo",
    Filters: []fugo.FileFilter{
        {Name: "Imágenes", Extensions: []string{"png", "jpg"}},
    },
})

savePath, err := ctx.ShowSaveDialog(fugo.FileDialogOptions{
    Title: "Guardar como",
    DefaultName: "archivo.txt",
})
```

Flutter maneja el diálogo nativo (vía `file_picker` o `file_selector`), envía el resultado a Go por gRPC.

### Clipboard

```go
// Go API
ctx.Clipboard().Write("texto a copiar")
text, err := ctx.Clipboard().Read()
```

Se usa `Clipboard` de Flutter Services (`ServicesBinding.instance.clipboard`).

---

## 5. Estimación de esfuerzo

| Componente | Complejidad | Tiempo estimado |
|-----------|------------|----------------|
| Window Manager (window_manager integración) | Media | 2 semanas |
| WindowDragArea widget | Baja | 1 semana |
| Process Supervisor (os/exec) | Media | 2 semanas |
| Signal handling (Linux/macOS/Windows) | Alta | 2 semanas |
| Heartbeat / Health check | Baja | 1 semana |
| Empaquetado Linux (binary único) | Media | 2 semanas |
| Empaquetado macOS (.app bundle) | Alta | 2 semanas |
| Empaquetado Windows (.exe) | Media | 2 semanas |
| File Picker integración | Media | 1 semana |
| Clipboard integración | Baja | 0.5 semanas |
| **Total Desktop** | — | **15.5 semanas** |

---

## 6. Entregables verificables

- [ ] Flutter se inicia como subproceso de Go en Linux
- [ ] Flutter se inicia como subproceso de Go en macOS
- [ ] Flutter se inicia como subproceso de Go en Windows
- [ ] Ctrl+C en terminal → ambos procesos terminan limpiamente
- [ ] Cerrar ventana Flutter → Go termina limpiamente
- [ ] Go panic → Flutter muere (sin zombie)
- [ ] Flutter crash → Go lo detecta y termina
- [ ] `fugo build` produce binario autocontenido en Linux
- [ ] `fugo build` produce .app en macOS
- [ ] `fugo build` produce .exe en Windows
- [ ] WindowDragArea funcional en modo frameless
- [ ] File picker: abrir/guardar diálogos nativos
- [ ] Clipboard: leer/escribir

---

## Referencias

- window_manager: <https://pub.dev/packages/window_manager>
- bitsdojo_window: <https://github.com/bitsdojo/bitsdojo_window>
- Go os/exec: <https://pkg.go.dev/os/exec>
- Go os/signal: <https://pkg.go.dev/os/signal>
- gRPC health checking: <https://grpc.io/docs/guides/health-checking/>
- Go embed: <https://pkg.go.dev/embed>
- Flutter desktop: <https://docs.flutter.dev/platform-integration/desktop>
- Linux prctl(2): `man 2 prctl`
- Linux unix(7): `man 7 unix`
