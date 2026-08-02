# Fugo Idiomático — la guía completa

> Este documento describe el estado **real** del código en el momento de escribirlo (no el ROADMAP aspiracional). Si algo aquí y el código difieren, gana el código — pero si eso pasa, este documento quedó desactualizado y hay que corregirlo.

## Índice

1. [Qué es Fugo, en una frase](#1-qué-es-fugo-en-una-frase)
2. [Filosofía: por qué está diseñado así](#2-filosofía-por-qué-está-diseñado-así)
3. [Arquitectura y flujo de datos](#3-arquitectura-y-flujo-de-datos)
4. [Primeros pasos: el CLI](#4-primeros-pasos-el-cli)
5. [Anatomía de un proyecto](#5-anatomía-de-un-proyecto)
6. [El paquete `fg`: construyendo UI](#6-el-paquete-fg-construyendo-ui)
7. [Catálogo completo de widgets](#7-catálogo-completo-de-widgets)
8. [Eventos: el formato de wire](#8-eventos-el-formato-de-wire)
9. [Theming](#9-theming)
10. [Estado idiomático: closures, Store[S], y por qué no hay BLoC](#10-estado-idiomático-closures-views-y-por-qué-no-hay-bloc)
11. [Navegación con Router](#11-navegación-con-router)
12. [Servicios del sistema operativo (host services)](#12-servicios-del-sistema-operativo-host-services)
13. [Ventana y overlays](#13-ventana-y-overlays)
14. [Testing](#14-testing)
15. [Plataformas soportadas](#15-plataformas-soportadas)
16. [Empaquetado, distribución e instalación](#16-empaquetado-distribución-e-instalación)
17. [Convenciones idiomáticas de código](#17-convenciones-idiomáticas-de-código)
18. [El protocolo interno (para quien quiera extender Fugo)](#18-el-protocolo-interno-para-quien-quiera-extender-fugo)
19. [Errores comunes y cómo evitarlos](#19-errores-comunes-y-cómo-evitarlos)

---

## 1. Qué es Fugo, en una frase

Fugo es un framework de **Server-Driven UI local para escritorio**: escribís toda la lógica, el estado y el ruteo de tu app **en Go puro**; un binario de Flutter precompilado actúa como una terminal de render tonta. Los dos procesos hablan por streaming bidireccional de gRPC (Unix Domain Socket en Linux/macOS, TCP en Windows). Go es la única fuente de verdad — Flutter no retiene lógica de negocio ni estado propio de la app.

No es Flutter con bindings de Go. No hay "widgets de Flutter que llamás desde Go". Vos construís un árbol de structs Go (`fg.Text(...)`, `fg.Button(...)`, ...), Fugo lo serializa a protobuf, lo manda por el socket, y del otro lado un registro de widgets Dart (`registry.dart`) decide qué widget real de Flutter dibujar por cada nodo.

## 2. Filosofía: por qué está diseñado así

- **Un solo lenguaje para todo.** No hay una capa de UI en Dart con lógica propia. Si mañana querés cambiar el motor de render (otro binario, otro protocolo), la lógica de tu app no se toca.
- **El árbol se construye una vez.** No es un modelo "rebuild from scratch" al estilo React/Flutter. Vos armás el árbol de widgets **una sola vez** al arrancar (`buildUI(ctx)`), guardás referencias a los widgets que te importan, y los handlers de eventos **mutan esos structs directamente** llamando después `ctx.Update()`.
- **Sin abstracciones que no necesitás.** No hay macros, no hay reflection para bindear widgets a campos, no hay un DSL de templates. Es Go llano: constructores encadenables (`fg.Text("hola").FontSize(20).Color(...)`) que devuelven `*fg.TextWidget`.
- **Material 3 por defecto, sin colores hardcodeados.** Los widgets no inyectan colores propios a menos que llames un setter explícito — el cliente construye un `ColorScheme.fromSeed` y todo lo no-seteado hereda ese esquema.

## 3. Arquitectura y flujo de datos

```
buildUI(ctx) ──> árbol fg.Widget (se construye UNA VEZ, se retiene)
                      │
   los handlers de evento mutan structs de widgets in-place, luego ctx.Update()
                      │  marca el scheduler como "dirty"
   engine.Scheduler (tick de 16ms / 60fps) ── si dirty ──> App.flush()
                      │
   fg.BuildTreeWithMerge: re-recorre el árbol retenido -> WidgetTree (proto)
                      │
   engine.Diff(oldTree, newTree) -> []Patch (CREATE/UPDATE/DELETE/REPLACE/REORDER)
                      │
   engine.Reconciler.SendPatches -> gRPC RenderStream -> Flutter
                      │
   Flutter aplica los patches a su mapa plano de WidgetNode, reconstruye vía WidgetRegistry
                      │
   interacción del usuario -> ClientEvent (debounced a 16ms, salvo onClick/onTap/onSubmit/
                               onLongPress/host/filedrop, que van inmediato) -> App.HandleEvent
```

**Modelo mental clave**: el árbol de widgets se construye **una vez** en `App.Run` y se **retiene**. Los handlers de eventos son closures de Go que mutan campos de structs directamente (ej. `counterText.SetText("5")`) y después llaman `ctx.Update()`. El scheduler vuelve a recorrer ese **mismo** árbol retenido cada frame si está sucio, re-marshalea las props, diffea contra el snapshot de protobuf anterior, y transmite solo los patches. **No** es un modelo de "reconstruir todo desde cero" tipo React.

### Paquetes del repo

| Paquete | Rol |
|---|---|
| `fugo` (raíz, `app.go`) | `App`, `Context`, ciclo de vida. `Run` construye el árbol + arranca el scheduler; `RunStandalone` además levanta el server gRPC, ajusta el GC y lanza Flutter. |
| `fg/` | La API declarativa de widgets — constructores **sin prefijo** (`fg.Text`, `fg.Container`, `fg.Button`, `fg.Router`...), no `NewText`. Cada uno devuelve un `*fg.XxxWidget` concreto con setters encadenables. |
| `style/` | Primitivas de estilo: `Color` (`style.Hex(...)`), `TextStyle`, `EdgeInsets` (`style.EdgeAll`), `Border`, pesos de fuente. |
| `config/` | Cargador sin dependencias de `fugo.toml`. Compartido por el runtime y el CLI. |
| `engine/` | `Diff`, `Reconciler` (envuelve el stream gRPC), `Scheduler` (coalesce de updates a un flush por frame). |
| `transport/` | Servidor gRPC (`StartServer`), health check, keepalive. UDS o TCP según la dirección. |
| `supervisor/` | Lanza y monitorea el subproceso de Flutter. |
| `cmd/fugo/` | El CLI: `init`, `run`, `build`, `doctor`, `autostart`, `widgets`, `generate`, `vet`, `fix`, `upgrade`. |
| `flutter_client/` | El cliente Dart. `registry.dart` mapea cada `WidgetType` a un widget real de Flutter. |

## 4. Primeros pasos: el CLI

```bash
fugo init myapp              # scaffolding: main.go + ui/ + fugo.toml + README + git init
cd myapp
fugo run                     # build + lanza el server Go + spawnea Flutter; hot-reload por defecto
fugo run --no-watch          # un solo build-and-run, sin watcher
fugo build                   # binario release + bundle de Flutter en dist/ + empaquetado Linux
fugo doctor                  # chequea toolchain; dentro de un proyecto valida coherencia completa
fugo doctor --fix            # repara lo auto-reparable (fugo.toml, git init, go mod tidy, fugovet -fix)
fugo autostart enable        # arranca la app al iniciar sesión
fugo widgets                 # explora el catálogo de fg y sus doc comments
fugo generate                # scaffolding de una pantalla o componente nuevo
fugo vet                     # analizador estático propio (fugovet)
fugo fix                     # auto-fix + gofumpt + reordenamiento de convención
fugo upgrade                 # autoactualiza el CLI (go install .../fugo@latest)
```

`fugo init` es interactivo (wizard) salvo que pases `--yes`/`-y` o no haya TTY; acepta `--theme`, `--organization`/`--org`, y `--template counter|app|showcase` (`counter` es el default).

`fugo run` usa `[server] addr` de `fugo.toml` como dirección por defecto (salvo `--addr`), y por defecto hace **hot reload**: observa `.go` con fsnotify y reconstruye/reconecta el server Go mientras la ventana de Flutter queda abierta (el estado en memoria se resetea entre reloads — no hay recuperación de estado completa todavía).

`fugo doctor` valida, dentro de un proyecto: estructura de archivos, coherencia `go.mod` ↔ `main.go`'s `<module>/ui` import ↔ `ui.Build` exportado, que el bundle de Flutter esté disponible, que la dirección configurada esté libre, `go build ./...`, y (nuevo) si `notify-send`/libnotify está disponible en Linux para `Context.Notifications()`.

## 5. Anatomía de un proyecto

```
myapp/
├─ main.go        # entrypoint: fija el theme, luego fugo.RunStandalone(fugo.ConfigOptions("fugo.toml"), ui.Build)
├─ ui/            # tus pantallas (package ui); ui.Build es el widget raíz
│  └─ home.go
├─ fugo.toml      # título/tamaño de ventana + dirección del server gRPC
├─ bin/           # builds de desarrollo (gitignored)
├─ dist/          # bundle de release (gitignored) — ver §16
├─ logs/          # fugo run → logs/run.log (gitignored)
├─ README.md
└─ .gitignore
```

`fugo.toml`:

```toml
name = "myapp"

[window]
title  = "My App"
width  = 800
height = 600

[server]
addr = "127.0.0.1:9510"

[app]
organization = "com.acme"   # opcional, solo si lo pasaste a fugo init
```

`config.ConfigOptions("fugo.toml")` lee este archivo, exporta `FUGO_ADDR` si hace falta, y devuelve un `fugo.AppOptions{Title, Width, Height}` listo para `RunStandalone`.

## 6. El paquete `fg`: construyendo UI

Todo widget implementa la interfaz `Widget` (bookkeeping interno) y se construye con una función **sin prefijo** que devuelve un puntero concreto:

```go
func Build(ctx *fugo.Context) fg.Widget {
    counter := 0
    counterText := fg.Text("0").FontSize(32)

    incBtn := fg.Button("+1").OnClick(func(fg.Event) {
        counter++
        counterText.SetText(strconv.Itoa(counter))
        ctx.Update()
    })

    return fg.Column(
        counterText,
        incBtn,
    )
}
```

Puntos clave:

- **`buildUI`/`ui.Build` corre una sola vez.** Guardá punteros a los widgets que un handler necesite mutar (closures capturan esas variables).
- **Cada setter devuelve el mismo widget** (`*fg.TextWidget`, etc.) para encadenar: `fg.Text("hola").FontSize(18).Color(style.Hex("#333"))`.
- **`ctx.Update()`** marca el árbol como sucio; el scheduler hace el resto (re-walk, diff, patch) en el próximo tick de 16ms. **`ctx.UpdateNow()`** salta la espera del tick para casos sensibles a latencia.
- **Componentes con estado** (alternativa a un closure de `buildUI`): implementá la interfaz `Component{ Render(ctx *Context) fg.Widget }` y pasala a `fugo.RunComponent(opts, component)`.

## 7. Catálogo completo de widgets

83 constructores en `fg/`, agrupados por `fugo widgets` así (más útil ejecutar ese comando localmente, que siempre refleja el código actual — esta es una foto):

### Layout
`Center`, `Column`, `Row`, `Stack`, `Container`, `Padding`/`PaddingAll`, `SizedBox`, `Expanded`, `Flexible`, `Spacer`, `Align`, `AspectRatio`, `ClipRRect`, `FittedBox`, `ConstrainedBox`, `FractionallySizedBox`, `Positioned`/`AnimatedPositioned`, `Wrap`, `Divider`/`VerticalDivider`, `Responsive` (elige entre hijos precompilados según el ancho disponible, resuelto 100% en el cliente vía `LayoutBuilder`, sin ida y vuelta a Go).

### Entrada / formularios
`TextField`, `Checkbox`, `Radio`, `Switch`, `Slider`/`RangeSlider`, `Dropdown`, `Autocomplete`, `Form` (agrega validación de múltiples campos vía `.AddField(validate)` + `.Validate() bool`), botones: `Button`/`FilledButton`/`FilledTonalButton`/`OutlinedButton`/`TextButton`/`ElevatedButton`/`IconButton`/`FloatingActionButton`/`SegmentedButton`.

### Scroll y listas
`ListView`, `GridView`, `ScrollView`, `Scrollbar`, `RefreshIndicator`, `PageView`, `SliverScaffold` (header colapsable + lista scrolleable).

### Navegación y estructura de pantalla
`Scaffold`, `AppBar` (título, leading, actions, `.Elevation`, `.ForegroundColor`, `.Bottom`), `NavigationBar`, `NavigationRail`, `Tabs`, `Router`, `Drawer` (vía `Scaffold`).

### Feedback y overlays
`Tooltip`, `Badge`, `ProgressCircular`/`ProgressLinear`, `Dismissible`, `Stepper`, `ExpansionTile`, `PopupMenuButton`.

### Datos y medios
`Card`, `DataTable` (con `.Sortable`/`.Selectable`), `Table`, `ListTile`, `CheckboxListTile`, `RadioListTile`, `SwitchListTile`, `Chip`, `CircleAvatar`, `RichText`, `Icon`, `Image`.

### Interacción avanzada
`GestureDetector` (tap, double-tap, long-press, pan, scale), `Draggable`/`DragTarget`, `WindowDragArea`.

### Animación y dibujo
`AnimatedContainer`, `AnimatedOpacity`, `AnimatedPositioned`, `Canvas` (superficie de dibujo acotada: `.Line`/`.Rect`/`.Circle`/`.Path`/`.FilledPath`, para sparklines/gauges, no drawing arbitrario).

### Accesibilidad
`Semantics` (anotación explícita para lectores de pantalla cuando Flutter no puede inferir una buena por su cuenta).

Cada constructor tiene su doc comment en el código fuente — `fugo widgets` los lista todos con su primera línea de documentación, siempre actualizado porque se regenera de `fg/*.go` (`make gen-widgets`).

## 8. Eventos: el formato de wire

Un `Event` (lo que recibe `Handle(event Event)`) tiene:

```go
type Event struct {
    NodeID    string
    EventType string
    Data      []byte
}
```

Convenciones (todas ad hoc pero consistentes en todo `fg/`):

| Tipo de valor | Formato de `Data` | Ejemplo |
|---|---|---|
| Booleano (Checkbox, Switch, listtiles) | `"1"`/`"0"` | `string(e.Data) == "1"` |
| Texto (TextField, Dropdown, Autocomplete) | string crudo | `string(e.Data)` |
| Float (Slider) | decimal como string | `strconv.ParseFloat(string(e.Data), 64)` |
| Rango (RangeSlider) | `"start,end"` con 2 decimales | `"20.50,80.00"` |
| Sort de columna (DataTable) | `"colIndex,asc"` | `"2,1"` |
| Selección de fila (DataTable) | `"rowIndex,selected"` | `"3,1"` |

Para **testear sin gRPC ni Flutter**, usá el testkit (§14): `fg.ClickEvent()`, `fg.BoolEvent(v)`, `fg.TextEvent(s)`, `fg.FloatEvent(v)`, `fg.RangeEvent(start, end)` construyen el `Event` sintético correcto sin que tengas que adivinar el formato.

## 9. Theming

```go
type Theme struct {
    Colors       ThemeColors      // 10 roles semánticos (Primary, Secondary, Background, Surface, Error, Success, OnPrimary, OnSurface, Muted, Border)
    Typography   ThemeTypography  // Family, Heading, Body, Caption, Weight
    Spacing      ThemeSpacing     // XS, SM, MD, LG, XL
    Radius       ThemeRadius      // SM, MD, LG
    Components   ThemeComponents  // CardRadius, CardElevation, ButtonRadius — overrides por tipo de widget
    Dark         bool             // brillo M3; ignorado si FollowSystem
    FollowSystem bool             // sigue el light/dark del SO en vivo (ThemeMode.system)
}
```

```go
fg.UseTheme(fg.DarkTheme())            // o fg.LightTheme() (default)
fg.CurrentTheme().Spacing.MD           // leer un token mientras armás UI
```

- `LightTheme()`/`DarkTheme()` son opinionados y ya vienen con `Typography`/`Spacing`/`Radius`/`Components` por defecto razonables.
- `Theme.Components` es el equivalente de Fugo a `CardTheme`/`ElevatedButtonThemeData` de Flutter: cambiá el radio/elevación de todas las Cards/Buttons desde un solo lugar (un valor en 0 en el widget individual cae al default del theme).
- `Theme.Typography.Family` se reenvía al cliente como `FUGO_THEME_FONT_FAMILY` y se aplica vía `ThemeData.fontFamily` — tiene que ser una fuente **ya instalada en el SO** por nombre; Fugo no empaqueta assets de fuentes.
- `Theme.FollowSystem = true` hace que el cliente siga el modo claro/oscuro del sistema operativo en vivo, sin que el proceso Go se entere del cambio (lo resuelve Flutter con `ThemeMode.system`).
- No hay theming por-widget más allá de Card/Button (no hay `ChipTheme`, `InputTheme`, `AppBarTheme` dedicados) ni tema Cupertino/adaptativo — es una decisión de diseño, no un olvido: Fugo es opinionadamente Material 3.

## 10. Estado idiomático: closures, `Store[S]`, y por qué no hay BLoC

### El modelo por defecto: closures + `ctx.Update()`

La forma normal de manejar estado en Fugo es **mutar directamente el struct de un widget** desde el handler de otro, usando clausuras de Go para capturar las referencias:

```go
total := fg.Text("$0.00")
items := []float64{}

addItemBtn := fg.Button("Agregar $10").OnClick(func(fg.Event) {
    items = append(items, 10)
    sum := 0.0
    for _, v := range items { sum += v }
    total.SetText(fmt.Sprintf("$%.2f", sum))
    ctx.Update()
})
```

No hay virtual-DOM al estilo React: el árbol se recorre de nuevo (rápido, es solo marshaling a protobuf) y se diffea posicionalmente contra el snapshot anterior — el patch que sale por el wire es mínimo.

**Esto alcanza para la enorme mayoría de apps.** Es explícito, no tiene magia, y es fácil de seguir un handler a la vez.

### Cuándo esto se queda corto

- Más de un puñado de widgets/pantallas necesitan reaccionar al mismo dato.
- Querés un único punto para inspeccionar "el estado actual" (útil en tests).
- Hay handlers concurrentes (un callback de `Context.Notifications()`/`Clipboard()`/`Files()` corre en el goroutine de transporte, no en el del scheduler) que tocan el mismo dato — los structs de widgets **no tienen mutex propio**, así que la seguridad de datos ahí es 100% responsabilidad tuya.

### `fg.Store[S]`: la respuesta opcional, no un Cubit

Fugo **no** copia el patrón BLoC/Cubit de Flutter (eventos → `emit()` → stream de snapshots inmutables → `BlocBuilder` reconstruye). Eso pelearía contra el árbol retenido: los widgets de Fugo ya SON las "vistas" mutables, y `ctx.Update()` ya dispara el diff — el trabajo que en Flutter hace `BlocBuilder.rebuild` acá ya lo hace el motor de reconciliación automáticamente.

Lo que sí agrega Fugo es un primitivo mínimo de **estado compartido con notificación**:

```go
type Store[S any] struct { /* ... */ }

func NewStore[S any](initial S) *Store[S]
func (s *Store[S]) Get() S
func (s *Store[S]) Set(next S)
func (s *Store[S]) Update(fn func(*S))       // estilo reducer: mutá una copia in-place
func (s *Store[S]) Subscribe(fn func(S))
```

Uso típico:

```go
type appState struct{ count int }

store := fg.NewStore(appState{})
countText := fg.Text("0")

store.Subscribe(func(s appState) {
    countText.SetText(strconv.Itoa(s.count))
    ctx.Update()
})

incBtn := fg.Button("+1").OnClick(func(fg.Event) {
    store.Update(func(s *appState) { s.count++ })
})
```

- **No reemplaza** `Context`/`Update()`/el árbol retenido — compone con ellos. Los suscriptores son closures comunes que sincronizan campos de widgets y llaman `ctx.Update()` ellos mismos.
- Es genérico (`S any`), sin reflection, con un solo mutex interno — resuelve tanto el problema de "una sola fuente de verdad inspeccionable" (`store.Get()` en un test) como el de concurrencia (`Set`/`Update` son thread-safe).
- Es **opt-in**: una app chica no lo necesita. Usalo cuando el dolor de wiring manual de closures sea real, no por adelantado.

## 11. Navegación con Router

```go
router := fg.Router(map[string]func() fg.Widget{
    "/":          HomeScreen,
    "/user/:id":  UserScreen,
}, "/")
```

- Soporta `:params` (ej. `/user/:id`), leídos con `ctx.Param("id")` dentro de la pantalla.
- `ctx.NavigateTo(route)` / `ctx.GoBack()` navegan; el `Router` anima la transición del lado del cliente (`AnimatedSwitcher`).
- `.OnBeforeLeave(func() bool)` permite vetar una navegación (ej. "¿descartar cambios sin guardar?" — devolver `false` cancela).
- Tiene su propio `WidgetType` (no reusa `CONTAINER` como placeholder).

## 12. Servicios del sistema operativo (host services)

Todos corren en el cliente y responden de forma asíncrona por un callback que se ejecuta en el goroutine de eventos (podés mutar widgets y llamar `ctx.Update()` desde ahí, igual que en cualquier handler).

```go
ctx.Clipboard().Write("texto")
ctx.Clipboard().Read(func(text string) { ... })

ctx.Files().Open(fg.FileDialog{Title: "Abrir", Extensions: []string{"png","jpg"}}, func(path string) { ... })
ctx.Files().Save(fg.FileDialog{DefaultName: "informe.csv"}, func(path string) { ... })

ctx.Notifications().Show("Título", "Cuerpo de la notificación")

ctx.RequestFocus(myTextField)   // hoy solo TextField implementa fg.Focusable

ctx.RegisterShortcuts(map[string]func(){
    "ctrl+s": func() { save() },
    "escape": func() { cancel() },
})

ctx.OnResize(func(width, height float64) { ... })
ctx.OnFileDrop(func(paths []string) { ... })   // archivos soltados desde el SO sobre la ventana
```

Notas honestas:
- `Files().Open`/`.Save` usan `file_selector` (no `file_picker` — ese no tiene implementación de Linux en absoluto, ver §15).
- `Notifications().Show` depende de libnotify/D-Bus en Linux (`fugo doctor` lo chequea) y del Notification Center nativo en Windows/macOS.
- No hay Focus/tab-order declarativo todavía — `RequestFocus` es la única primitiva, y solo para `TextField`.

## 13. Ventana y overlays

```go
ctx.Window().SetTitle("Nuevo título")
ctx.Window().SetSize(1024, 768)
ctx.Window().Minimize() / .Maximize() / .Center() / .SetFullScreen(true)

ctx.ShowSnackBar("Guardado")
ctx.ShowDialog("Confirmar", "¿Seguro que querés continuar?")       // un solo botón OK
ctx.ShowBottomSheet("Detalles", "Más información acá")

ctx.PickDate(func(date string) { ... })   // "YYYY-MM-DD"
ctx.PickTime(func(t string) { ... })      // "HH:MM" 24hs
```

> Nota de precisión: el proto y el cliente Dart ya soportan botones de acción custom en diálogos/bottom sheets (`OverlayCommand.actions`), pero **no existe todavía** un método público `ShowDialogActions`/`ShowBottomSheetActions` en `Context` que lo exponga — es un cabo suelto real, no una limitación de diseño. Si lo necesitás hoy, tocás `sendOverlay` directamente o pedís que se agregue la API pública.

## 14. Testing

No hace falta gRPC, Reconciler ni el proceso de Flutter para testear una UI de Fugo:

```go
tree, widgetsByID := fg.BuildTree(ui.Build(ctx))  // asigna ids, una sola vez

btn := someKnownWidgetReference // guardaste la referencia al construir el árbol
btn.Handle(fg.ClickEvent())

// o por id:
w := widgetsByID[5]
w.Handle(fg.BoolEvent(true))

// asertá sobre campos exportados del propio widget o de lo que el handler mutó
```

Constructores sintéticos de `Event` disponibles: `ClickEvent()`, `BoolEvent(v)`, `TextEvent(s)`, `FloatEvent(v)`, `RangeEvent(start, end)` — cada uno documenta qué widgets lo usan y por qué (ver `fg/testkit.go` y `fg/testkit_example_test.go`).

## 15. Plataformas soportadas

| Plataforma | Estado |
|---|---|
| Linux (X11) | ✅ Soportado |
| Linux (Wayland — Hyprland, Sway, GNOME, ...) | ✅ Funciona vía **XWayland** — probado corriendo `fugo-spike` en una sesión Hyprland/Wayland real (no es una inferencia del código: `flutter_client/linux/runner/my_application.cc`, la plantilla estándar de `flutter create`, solo detecta X11 explícitamente y asume que Wayland "probablemente funcione", sin lógica propia de Fugo para eso). El embedder Linux de Flutter todavía no es nativamente Wayland, así que esto depende de que el compositor corra XWayland. |
| Windows | ✅ Soportado (transporte TCP en vez de UDS) |
| macOS | No testeado activamente, debería compilar (mismo transporte UDS que Linux) |
| Android / iOS | ❌ **Fuera de alcance**, no es un "todavía no" — el modelo (Go como subproceso lanzado por Flutter, hablando por Unix socket) no encaja en un sandbox móvil. Soportarlo requeriría embeber el runtime de Go in-process (`gomobile`/cgo) en vez de spawnearlo — una arquitectura distinta, no un cambio incremental. |

## 16. Empaquetado, distribución e instalación

`fugo build` genera en `dist/`:

- El binario release (`go build -ldflags="-s -w"`).
- `fugo.toml` (si existe) copiado al lado.
- El bundle precompilado de Flutter (descargado automáticamente según la versión del CLI, o tomado de un build local).
- **En Linux**, además: un `.desktop` (entrada XDG estándar — editá `Icon=` para apuntar a tu propio ícono), un `install.sh` (instala para el usuario actual sin root: binario a `~/.local/share/<app>`, symlink en `~/.local/bin`, entrada en `~/.local/share/applications`), y una plantilla `PKGBUILD` para publicar en AUR.

`fugo autostart enable`/`disable` registra (o quita) el binario de `dist/` para que arranque al iniciar sesión:
- **Linux**: entrada XDG autostart en `~/.config/autostart/`. Ojo: GNOME/KDE/XFCE la procesan solos; compositores Wayland "tiling" sin session manager completo (Hyprland, Sway) generalmente **no** — necesitás algo como `dex` invocado desde tu config (`exec-once = dex ~/.config/autostart -a` en Hyprland).
- **Windows**: clave `Run` en `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` (vía `reg.exe`, sin dependencia extra de Go).

No hay (todavía) generación automática de AppImage/`.deb`/`.rpm`/`.msi`/instalador gráfico — esos pasos son manuales si los necesitás; `install.sh` + `PKGBUILD` son el punto de partida.

## 17. Convenciones idiomáticas de código

- Formateador: **`gofumpt`**, no `gofmt` — `gofumpt -w .` antes de commitear.
- Constructores de widgets **sin prefijo** (`fg.Text`, no `fg.NewText`), devolviendo un puntero concreto, con setters que devuelven `self` para encadenar.
- `*Set bool` flags para distinguir "el usuario nunca llamó este setter" (cae al default del theme) de "el usuario explícitamente puso este valor" — patrón usado en `Container.BgColor`, `Button.BgColor`, `AppBar.BgColor`/`ForegroundColor`, etc.
- El versionado de Fugo sigue exactamente la versión de Flutter (`<flutter-version>-fugo.<build>`) — nunca se edita `VERSION`/`FLUTTER_VERSION`/`CHANGELOG.md` a mano, se usa `make release`/`make release-flutter`.
- Agregar un widget nuevo toca **4 lugares**: el proto (`WidgetType` + mensaje `*Props`), `make proto` (regenera Go + Dart), el widget en `fg/`, y el caso en `flutter_client/lib/registry.dart`.

## 18. El protocolo interno (para quien quiera extender Fugo)

`WidgetNode.props` es un campo `bytes` que contiene un mensaje protobuf **por tipo de widget** (`TextProps`, `ButtonProps`, ...) — protobuf anidado dentro de protobuf, marshaleado con `proto.Marshal` en Go y decodeado con `*.fromBuffer(node.props)` en Dart. Todo el `RenderPayload` es un mensaje protobuf normal de gRPC.

`transport/proto/fugo/v1/fugo.proto` es la fuente única. `make proto` corre `protoc` y regenera **tanto** los bindings Go (`*.pb.go`, committeados al repo — así `go install`/un checkout limpio compilan sin `protoc`) **como** los bindings Dart (gitignored, se regeneran con `flutter build`).

## 19. Errores comunes y cómo evitarlos

- **"Muté un widget pero no pasa nada"** → te olvidaste `ctx.Update()` (o `ctx.UpdateNow()` si necesitás que se refleje antes del próximo tick de 16ms).
- **"El campo del proto existe pero no hace nada"** → puede que el setter en Go nunca haya sido escrito (pasó con `TextField.SetError`/`Dropdown.SetError`: el campo `error_text` existía en el proto y el cliente Dart ya lo renderizaba, pero el widget Go nunca lo exponía ni lo marshaleaba). Ante la duda, seguí la cadena completa: proto → Go widget → Dart registry.
- **"Uso una fuente/ícono personalizado y no aparece"** → las fuentes son por nombre de sistema (`Theme.Typography.Family`), no assets embebidos; los íconos son el set base de Material (`fg.Icons.X`), no arbitrarios.
- **"Necesito que dos pantallas compartan estado"** → no hardcodees el wiring de closures entre archivos separados si se vuelve inmanejable; usá `fg.Store[S]` (§10).
- **"Mi app no arranca en Wayland"** → confirmá que XWayland esté corriendo (`ps aux | grep -i xwayland`); es lo que usa el embedder de Flutter hoy.
