# 07 — API Go: Superficie Pública

## Alcance

Define la API pública que el desarrollador Go usa para construir aplicaciones Fugo. Es la única superficie que el desarrollador toca. Comprende:

- **`fugo`**: Inicialización del framework, App, Context, ciclo de vida
- **`fugo/ui`**: Widgets declarativos (Container, Text, Button, Row, ...)
- **`fugo/style`**: Primitivas de estilo (Font, Color, Border, Shadow, ...)

[Ver SPEC.md:§5 para la visión original de la API]
[Ver 02_ARQUITECTURA.md para cómo esta API alimenta el Core Engine]
[Ver 04_FLUTTER_CLIENT.md:§3 para el mapping Flutter correspondiente]

---

## 1. Paquete `fugo` — Inicialización y ciclo de vida

### App y Context

```go
package fugo

type App struct { ... }

type AppOptions struct {
    Title  string
    Width  int
    Height int
}

func NewApp(opts AppOptions) *App

// Run ejecuta el closure de UI y bloquea hasta que la app cierre.
// El closure se re-ejecuta cada vez que ctx.Update() es llamado.
func (a *App) Run(render func(ctx *Context) ui.Widget)
```

**`Context`** es el puente entre el desarrollador y el motor:

```go
type Context struct { ... }

// Update marca el contexto como sucio. En el próximo tick del scheduler,
// se re-ejecuta el closure de Run, se diffea, y se envía a Flutter.
func (c *Context) Update()

// NavigateTo cambia la ruta actual. El Router widget reacciona.
func (c *Context) NavigateTo(path string)

// NavigateBack retorna a la ruta anterior
func (c *Context) NavigateBack()

// Window devuelve el controlador de ventana
func (c *Context) Window() *WindowController
```

### Component Model (opcional, para state management avanzado)

```go
// Component es una interfaz opcional para componentes con estado.
// Si un widget implementa Component, Fugo gestiona su ciclo de vida.
type Component interface {
    // Render retorna el árbol de UI basado en el estado actual.
    Render(ctx *Context) ui.Widget
}
```

Ejemplo de uso:

```go
type Counter struct {
    count int
}

func (c *Counter) Render(ctx *fugo.Context) ui.Widget {
    return ui.Column(
        ui.Text(fmt.Sprintf("Count: %d", c.count)).FontSize(32),
        ui.Button("Increment").OnClick(func(e ui.Event) {
            c.count++
            ctx.Update()
        }),
    )
}
```

[Ver DX_IDEA.md y DX_OPINIONATED.md en docs/ para la filosofía completa del Component Model]

---

## 2. Paquete `fugo/ui` — Widgets

### Principios de diseño

- **Builder pattern**: Cada widget es inmutable. Los métodos de configuración retornan una copia modificada.
- **Encadenamiento fluido**: Los métodos se encadenan para legibilidad.
- **Composición sobre herencia**: No hay jerarquía de widgets. Todo se compone con Containers.
- **Mapping 1:1 con Flutter**: Cada widget Go mapea a exactamente un widget Flutter.

### Widgets de estructura

```go
package ui

// Container — el widget fundamental. Envuelve un hijo con padding, color, bordes, etc.
func Container(child Widget) *ContainerWidget

func (c *ContainerWidget) Padding(pixels float64) *ContainerWidget
func (c *ContainerWidget) PaddingXY(x, y float64) *ContainerWidget
func (c *ContainerWidget) PaddingOnly(top, right, bottom, left float64) *ContainerWidget
func (c *ContainerWidget) Margin(pixels float64) *ContainerWidget
func (c *ContainerWidget) BgColor(hex string) *ContainerWidget
func (c *ContainerWidget) BorderRadius(radius float64) *ContainerWidget
func (c *ContainerWidget) Border(width float64, color string) *ContainerWidget
func (c *ContainerWidget) Width(pixels float64) *ContainerWidget
func (c *ContainerWidget) Height(pixels float64) *ContainerWidget
func (c *ContainerWidget) MinWidth(pixels float64) *ContainerWidget
func (c *ContainerWidget) MaxWidth(pixels float64) *ContainerWidget
func (c *ContainerWidget) Align(alignment Alignment) *ContainerWidget
func (c *ContainerWidget) Shadow(s Shadow) *ContainerWidget
func (c *ContainerWidget) Style(s *style.Style) *ContainerWidget
func (c *ContainerWidget) Fill() *ContainerWidget  // Expandirse al padre
func (c *ContainerWidget) OnClick(handler func(Event)) *ContainerWidget

// Row — hijos en dirección horizontal
func Row(children ...Widget) *RowWidget
func (r *RowWidget) MainAxisAlignment(align MainAxisAlignment) *RowWidget
func (r *RowWidget) CrossAxisAlignment(align CrossAxisAlignment) *RowWidget
func (r *RowWidget) WithGap(pixels float64) *RowWidget

// Column — hijos en dirección vertical
func Column(children ...Widget) *ColumnWidget
func (c *ColumnWidget) MainAxisAlignment(align MainAxisAlignment) *ColumnWidget
func (c *ColumnWidget) CrossAxisAlignment(align CrossAxisAlignment) *ColumnWidget
func (c *ColumnWidget) WithGap(pixels float64) *ColumnWidget
func (c *ColumnWidget) Align(alignment Alignment) *ColumnWidget

// Stack — hijos apilados en Z
func Stack(children ...Widget) *StackWidget

// Positioned — posicionamiento absoluto dentro de Stack
func Positioned(child Widget) *PositionedWidget
func (p *PositionedWidget) Top(pixels float64) *PositionedWidget
func (p *PositionedWidget) Right(pixels float64) *PositionedWidget
func (p *PositionedWidget) Bottom(pixels float64) *PositionedWidget
func (p *PositionedWidget) Left(pixels float64) *PositionedWidget

// Expanded — expandir hijo dentro de Row/Column
func Expanded(child Widget) *ExpandedWidget
func (e *ExpandedWidget) Flex(factor int) *ExpandedWidget

// Center — centrar hijo
func Center(child Widget) *CenterWidget

// Padding — padding explícito
func Padding(child Widget, pixels float64) *PaddingWidget

// SizedBox — caja de tamaño fijo
func SizedBox(width, height float64) *SizedBoxWidget

// Wrap — layout que envuelve hijos a la siguiente línea
func Wrap(children ...Widget) *WrapWidget
func (w *WrapWidget) Direction(direction Axis) *WrapWidget
func (w *WrapWidget) Spacing(pixels float64) *WrapWidget
```

### Widgets de contenido

```go
// Text — texto renderizado
func Text(value string) *TextWidget
func (t *TextWidget) FontSize(pixels float64) *TextWidget
func (t *TextWidget) Font(f *style.Font) *TextWidget
func (t *TextWidget) Color(hex string) *TextWidget
func (t *TextWidget) Weight(weight style.FontWeight) *TextWidget
func (t *TextWidget) LetterSpacing(pixels float64) *TextWidget
func (t *TextWidget) LineHeight(multiplier float64) *TextWidget
func (t *TextWidget) Align(align TextAlign) *TextWidget
func (t *TextWidget) Overflow(overflow TextOverflow) *TextWidget
func (t *TextWidget) MaxLines(n int) *TextWidget
func (t *TextWidget) SetText(value string)  // Mutación en caliente (útil con ctx.Update)

// Image — imagen desde URL, asset, o bytes
func ImageURL(url string) *ImageWidget
func ImageAsset(path string) *ImageWidget
func ImageBytes(data []byte) *ImageWidget
func (i *ImageWidget) Fit(fit BoxFit) *ImageWidget
func (i *ImageWidget) Width(pixels float64) *ImageWidget
func (i *ImageWidget) Height(pixels float64) *ImageWidget

// Icon — icono SVG o material
func Icon(name string) *IconWidget
func (i *IconWidget) Size(pixels float64) *IconWidget
func (i *IconWidget) Color(hex string) *IconWidget

// Divider — línea divisoria
func Divider() *DividerWidget
func (d *DividerWidget) Thickness(pixels float64) *DividerWidget
func (d *DividerWidget) Color(hex string) *DividerWidget
```

### Widgets de entrada

```go
// Button — botón clickeable
func Button(label string) *ButtonWidget
func (b *ButtonWidget) OnClick(handler func(Event)) *ButtonWidget
func (b *ButtonWidget) Disabled(disabled bool) *ButtonWidget
func (b *ButtonWidget) Padding(pixels float64) *ButtonWidget
func (b *ButtonWidget) BorderRadius(pixels float64) *ButtonWidget
func (b *ButtonWidget) Style(s *style.Style) *ButtonWidget

// TextField — campo de entrada de texto
func TextField() *TextFieldWidget
func (t *TextFieldWidget) Placeholder(text string) *TextFieldWidget
func (t *TextFieldWidget) Value(text string) *TextFieldWidget
func (t *TextFieldWidget) Obscure(obscure bool) *TextFieldWidget  // Password
func (t *TextFieldWidget) MaxLines(n int) *TextFieldWidget
func (t *TextFieldWidget) OnChange(handler func(value string)) *TextFieldWidget
func (t *TextFieldWidget) OnSubmit(handler func(value string)) *TextFieldWidget

// Checkbox — caja de verificación
func Checkbox(checked bool) *CheckboxWidget
func (c *CheckboxWidget) OnChange(handler func(checked bool)) *CheckboxWidget

// Switch — interruptor
func Switch(active bool) *SwitchWidget
func (s *SwitchWidget) OnChange(handler func(active bool)) *SwitchWidget

// Slider — deslizador
func Slider(value, min, max float64) *SliderWidget
func (s *SliderWidget) Divisions(n int) *SliderWidget
func (s *SliderWidget) OnChange(handler func(value float64)) *SliderWidget
func (s *SliderWidget) OnChangeEnd(handler func(value float64)) *SliderWidget
```

### Widgets de listas y scroll

```go
// ListView — lista virtualizada (infinita)
func ListView(children ...Widget) *ListViewWidget
func (l *ListViewWidget) ScrollDirection(direction Axis) *ListViewWidget
func (l *ListViewWidget) Builder(count int, builder func(index int) Widget) *ListViewWidget

// GridView — grilla de elementos
func GridView(children ...Widget) *GridViewWidget
func (g *GridViewWidget) CrossAxisCount(n int) *GridViewWidget
func (g *GridViewWidget) AspectRatio(ratio float64) *GridViewWidget
```

### Widgets de navegación

```go
// Router — enrutador integrado
func Router(routes ...Route) *RouterWidget
func (r *RouterWidget) InitialRoute(path string) *RouterWidget

type Route struct {
    Path string
    View interface{}  // Widget o func(RouteArgs) Widget
}

type RouteArgs struct { ... }
func (a RouteArgs) Get(key string) string
```

### Widgets animados

```go
// AnimatedContainer — container con transiciones implícitas
func AnimatedContainer(child Widget) *AnimatedContainerWidget
func (a *AnimatedContainerWidget) Duration(ms int) *AnimatedContainerWidget
func (a *AnimatedContainerWidget) Curve(curve AnimationCurve) *AnimatedContainerWidget
// Hereda todos los métodos de Container (Width, BgColor, etc.)

// AnimatedOpacity — fade in/out
func AnimatedOpacity(child Widget, opacity float64) *AnimatedOpacityWidget
func (a *AnimatedOpacityWidget) Duration(ms int) *AnimatedOpacityWidget
```

---

## 3. Paquete `fugo/style` — Primitivas de Estilo

```go
package style

// Style agrupa propiedades visuales
type Style struct { ... }

func New(properties ...StyleProperty) *Style

// Propiedades de estilo
func BgColor(hex string) StyleProperty
func TextColor(hex string) StyleProperty
func FontSize(pixels float64) StyleProperty
func FontWeight(weight int) StyleProperty
func Padding(pixels float64) StyleProperty
func Margin(pixels float64) StyleProperty
func BorderRadius(pixels float64) StyleProperty
func Border(width float64, color string) StyleProperty

// Font
type Font struct { ... }
func Font(family string, weight FontWeight) *Font

// FontWeight
type FontWeight int
const (
    WeightThin       FontWeight = 100
    WeightLight      FontWeight = 300
    WeightRegular    FontWeight = 400
    WeightMedium     FontWeight = 500
    WeightBold       FontWeight = 700
    WeightBlack      FontWeight = 900
)

// Color — utilidades
func Hex(hex string) Color
func RGBA(r, g, b uint8, a float64) Color

// Shadow
type Shadow struct {
    Color  string
    OffsetX, OffsetY float64
    Blur   float64
    Spread float64
}

// Alignment
type Alignment int
const (
    AlignTopLeft      Alignment = 0
    AlignTopCenter    Alignment = 1
    AlignTopRight     Alignment = 2
    AlignCenterLeft   Alignment = 3
    AlignCenter       Alignment = 4
    AlignCenterRight  Alignment = 5
    AlignBottomLeft   Alignment = 6
    AlignBottomCenter Alignment = 7
    AlignBottomRight  Alignment = 8
)

// Animation
type AnimationCurve int
const (
    CurveLinear      AnimationCurve = 0
    CurveEaseIn      AnimationCurve = 1
    CurveEaseOut     AnimationCurve = 2
    CurveEaseInOut   AnimationCurve = 3
    CurveBounce      AnimationCurve = 4
    CurveElastic     AnimationCurve = 5
)
```

---

## 4. Sistema de eventos

```go
package ui

type Event struct {
    NodeID    string
    Type      string
    Timestamp int64
    Data      []byte  // FlatBuffer con datos específicos del evento
}

// Eventos específicos (wrappers alrededor de Event)
type ClickEvent = Event
type HoverEvent = Event
type ChangeEvent = Event
type DragEvent struct {
    Event
    DeltaX float64
    DeltaY float64
}
```

### Mapeo de eventos Flutter → Go

| Evento Flutter | Manejador Go |
|---------------|-------------|
| `onTap` | `OnClick(func(Event))` |
| `onDoubleTap` | `OnDoubleClick(func(Event))` |
| `onLongPress` | `OnLongPress(func(Event))` |
| `onHover` (MouseRegion) | `OnHover(func(Event))` |
| `onPanUpdate` | `OnDrag(func(DragEvent))` |
| `onChanged` (TextField) | `OnChange(func(string))` |

---

## 5. Sistema de Keys

Las keys proporcionan identidad estable a widgets entre frames, esencial para:
- Listas dinámicas (reordenamiento sin perder estado)
- Preservar estado de widgets (scroll position, foco)
- Optimizar diffing (identidad en vez de posición)

```go
// Key asigna una identidad estable
func (w *ContainerWidget) Key(key string) *ContainerWidget
func (w *TextWidget) Key(key string) *TextWidget
// ... disponible en todos los widgets
```

Uso:

```go
for i, item := range items {
    ui.Container(ui.Text(item.Name)).Key(fmt.Sprintf("item_%d", item.ID))
}
```

---

## 6. Controlador de ventana

```go
type WindowController struct { ... }

func (w *WindowController) Minimize()
func (w *WindowController) Maximize()
func (w *WindowController) Close()
func (w *WindowController) SetTitle(title string)
func (w *WindowController) SetSize(width, height int)
func (w *WindowController) Center()
func (w *WindowController) SetFullScreen(fullscreen bool)
func (w *WindowController) SetFrameless(frameless bool)
```

Acceso desde el closure de `app.Run`:

```go
app.Run(func(ctx *fugo.Context) ui.Widget {
    ctx.Window().SetTitle("Mi Aplicación")
    ctx.Window().Center()
    return ...
})
```

---

## 7. Constantes y enums

```go
package ui

type Axis int
const (
    AxisHorizontal Axis = 0
    AxisVertical   Axis = 1
)

type MainAxisAlignment int
const (
    MainStart    MainAxisAlignment = 0
    MainEnd      MainAxisAlignment = 1
    MainCenter   MainAxisAlignment = 2
    MainSpaceBetween MainAxisAlignment = 3
    MainSpaceAround  MainAxisAlignment = 4
    MainSpaceEvenly  MainAxisAlignment = 5
)

type CrossAxisAlignment int
const (
    CrossStart    CrossAxisAlignment = 0
    CrossEnd      CrossAxisAlignment = 1
    CrossCenter   CrossAxisAlignment = 2
    CrossStretch  CrossAxisAlignment = 3
)

type TextAlign int
const (
    TextAlignLeft    TextAlign = 0
    TextAlignRight   TextAlign = 1
    TextAlignCenter  TextAlign = 2
    TextAlignJustify TextAlign = 3
)

type TextOverflow int
const (
    OverflowClip     TextOverflow = 0
    OverflowEllipsis TextOverflow = 1
    OverflowFade     TextOverflow = 2
)

type BoxFit int
const (
    FitContain BoxFit = 0
    FitCover   BoxFit = 1
    FitFill    BoxFit = 2
    FitNone    BoxFit = 3
)
```

---

## 8. Ejemplo completo

```go
package main

import (
    "fmt"
    "github.com/sazardev/fugo"
    "github.com/sazardev/fugo/ui"
    "github.com/sazardev/fugo/style"
)

func main() {
    app := fugo.NewApp(fugo.AppOptions{
        Title:  "Fugo Counter",
        Width:  800,
        Height: 600,
    })

    baseFont := style.Font("Inter", style.WeightBold)
    darkTheme := style.New(
        style.BgColor("#121212"),
        style.TextColor("#FFFFFF"),
    )

    app.Run(func(ctx *fugo.Context) ui.Widget {
        counter := 0
        counterText := ui.Text("0").
            FontSize(48).
            Font(baseFont).
            Style(darkTheme)

        incrementBtn := ui.Button("Increment").
            OnClick(func(e ui.Event) {
                counter++
                counterText.SetText(fmt.Sprint(counter))
                ctx.Update()
            }).
            Padding(16).
            BorderRadius(4)

        return ui.Container(
            ui.Center(
                ui.Column(
                    counterText,
                    incrementBtn,
                ).WithGap(24),
            ),
        ).Style(darkTheme).Fill()
    })
}
```

---

## 9. Estimación de esfuerzo

| Componente | Complejidad | Tiempo estimado |
|-----------|------------|----------------|
| `fugo` (App, Context, ciclo de vida) | Media | 2 semanas |
| `fugo/ui` — Widgets de estructura (8 widgets) | Media | 3 semanas |
| `fugo/ui` — Widgets de contenido (5 widgets) | Media | 2 semanas |
| `fugo/ui` — Widgets de entrada (6 widgets) | Alta | 3 semanas |
| `fugo/ui` — Listas y scroll | Media | 2 semanas |
| `fugo/ui` — Router y navegación | Alta | 2 semanas |
| `fugo/ui` — Widgets animados (3 widgets) | Media | 2 semanas |
| `fugo/style` — Sistema de estilos | Media | 2 semanas |
| Sistema de eventos | Media | 1 semana |
| Sistema de Keys | Baja | 1 semana |
| Window Controller | Media | 1 semana |
| Documentación de API (godoc) | Baja | 2 semanas |
| **Total API Go** | — | **23 semanas** |

---

## 10. Entregables verificables

- [ ] API completa compila sin errores
- [ ] Cada widget Go tiene su builder Flutter correspondiente registrado
- [ ] Ejemplo Counter funcional end-to-end
- [ ] Ejemplo Router con 3 páginas funcional
- [ ] Ejemplo ListView con 1000 items virtualizados
- [ ] Tests unitarios para cada widget (verificación de estructura generada, no renderizado)
- [ ] Godoc completo para todos los símbolos exportados
- [ ] No hay imports circulares entre paquetes

---

## Referencias

- SPEC.md: `../SPEC.md:§5`
- DX_IDEA.md: `../docs/DX_IDEA.md`
- DX_OPINIONATED.md: `../docs/DX_OPINIONATED.md`
- ANEXUS_1.md: `../docs/ANEXUS_1.md`
- ANEXUS_FLUTTER.md: `../docs/ANEXUS_FLUTTER.md`
- Flutter widget catalog: <https://docs.flutter.dev/ui/widgets>
- Flutter layout: <https://docs.flutter.dev/ui/layout>
- Effective Go (naming conventions): <https://go.dev/doc/effective_go>
