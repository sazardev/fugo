# ANNEX-DX.md: Fugo - Developer Experience & Opinionated State Management

## 1. The Fugo Philosophy: An Opinionated DX

Fugo is not a sandbox; it is a highly structured, opinionated framework. We believe that unlimited choices in state management and architecture lead to fragmented ecosystems and unmaintainable codebases.

In Fugo, **Go is the absolute source of truth**. Flutter is strictly a transient visualization layer. By enforcing a single, unified way to handle state, routing, and component lifecycles, Fugo guarantees that a project written by one developer is instantly understandable by another.

The developer experience (DX) is engineered to be clean, type-safe, and deeply integrated with Go's standard idioms.

---

## 2. Opinionated State Management

Fugo does not offer ten different state management libraries. It provides exactly one: **The Fugo Component Model**.

State in Fugo lives entirely in Go memory. There is no "shared state" with the Flutter client. Data flows in one direction, and events bubble up.

### The Component Interface

A component in Fugo is simply a Go `struct` that implements a `Render()` method. State mutation is explicit and strictly localized.

```go
package components

import (
	"fmt"
	"github.com/fugo-ui/fugo/ui"
)

// 1. Define the State
type Counter struct {
	count int
	limit int
}

// 2. Initialize the Component
func NewCounter(limit int) *Counter {
	return &Counter{count: 0, limit: limit}
}

// 3. Render the UI based strictly on current state
func (c *Counter) Render(ctx *ui.Context) ui.Widget {
	return ui.Column(
		ui.Text(fmt.Sprintf("Current count: %d", c.count)).
			FontSize(24),

		ui.Button("Add").
			Disabled(c.count >= c.limit).
			OnClick(func(e ui.Event) {
				c.count++
				ctx.MarkDirty(c) // Explicitly tells Fugo to re-evaluate this component
			}),
	).Align(ui.Center)
}

```

**The Rule:** You cannot mutate the UI directly (e.g., `button.SetText("New")`). You mutate the state (`c.count++`) and tell the context to re-render (`ctx.MarkDirty(c)`). The diffing engine handles the rest.

---

## 3. Scalable Routing (Pages & Navigation)

Desktop applications grow fast. Fugo enforces a file-based routing mental model translated into Go structs, making it trivial to scale from a single-window utility to a massive, multi-page application.

Fugo includes a built-in, opinionated `Router` widget.

```go
package main

import (
	"github.com/fugo-ui/fugo"
	"github.com/fugo-ui/fugo/ui"
	"myapp/pages"
)

func main() {
	app := fugo.NewApp(fugo.AppOptions{Title: "Enterprise App"})

	app.Run(func(ctx *fugo.Context) ui.Widget {
		// The router is the root widget. It handles history, push, and pop.
		return ui.Router(
			ui.Route{
				Path: "/",
				View: pages.DashboardPage(),
			},
			ui.Route{
				Path: "/settings",
				View: pages.SettingsPage(),
			},
			ui.Route{
				Path: "/users/:id",
				View: func(args ui.RouteArgs) ui.Widget {
					return pages.UserProfilePage(args.Get("id"))
				},
			},
		).InitialRoute("/")
	})
}

```

Navigation is triggered via the context: `ctx.NavigateTo("/settings")`. The Flutter client smoothly transitions the views using native desktop animations automatically.

---

## 4. Unleashing Flutter's Power (The Go Wrapper)

Fugo aims to expose the immense power of Flutter's rendering engine—its complex layouts, flex constraints, and buttery-smooth animations—without forcing the developer to write Dart.

### Layouts & Flexbox

Instead of abstracting layouts into confusing new concepts, Fugo maps directly to Flutter's highly successful layout model (Rows, Columns, Flex, Expanded, Padding).

```go
ui.Row(
	ui.Expanded(
		ui.Container(ui.Text("Left Sidebar")).BgColor("#2A2A2A"),
	).Flex(1), // Takes 1/4 of the space

	ui.Expanded(
		ui.Container(ui.Text("Main Content")).BgColor("#121212"),
	).Flex(3), // Takes 3/4 of the space
)

```

### Implicit Animations

Flutter shines at UI animations. Fugo leverages "Implicit Animations" to make the DX magical. Developers do not write animation controllers or ticker providers in Go. They simply use animated variants of widgets, and Flutter interpolates the changes.

```go
// State variables
var isExpanded bool = false
var color string = "#FF0000"

// Inside the Render method:
ui.AnimatedContainer(
	ui.Text("Click me to morph"),
).
	Width(func() float64 { if isExpanded { return 400 } return 100 }()).
	BgColor(color).
	Duration(300). // 300ms transition
	Curve(ui.CurveEaseInOut).
	OnClick(func(e ui.Event) {
		isExpanded = !isExpanded
		color = "#00FF00"
		ctx.MarkDirty(c)
	})

```

When `MarkDirty` is called, Go calculates the new properties (width: 400, color: Green) and sends them to Flutter. Flutter automatically animates the transition from the old state to the new state over 300ms.

---

## 5. Customization: The "Blank Slate" Strategy

While Fugo is opinionated about _how_ data flows, it is entirely un-opinionated about _how_ things look.

Fugo does not ship with a bloated UI library mimicking an OS. It ships with raw primitives (`Container`, `Text`, `Shape`, `Icon`). If a team wants to build a Brutalist UI, a Neumorphic UI, or a macOS clone, they compose these primitives into their own internal Go packages.

```go
// Custom component built by a developer team
package designsystem

import "github.com/fugo-ui/fugo/ui"

// A strict, opinionated button for the company's internal tools
func PrimaryButton(label string, action func()) ui.Widget {
	return ui.Container(
		ui.Text(label).
			Color("#FFFFFF").
			Weight(ui.WeightBold).
			Tracking(1.5), // Letter spacing
	).
	BgColor("#0055FF").
	PaddingXY(24, 12).
	Cursor(ui.CursorPointer).
	OnClick(func(e ui.Event) { action() })
}

```

By enforcing strict state management while providing absolute visual freedom, Fugo delivers a DX that is deeply satisfying, predictable, and remarkably fast to iterate on.
