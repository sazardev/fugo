package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ButtonWidget is a Material button. Build one with a variant constructor:
// Button/FilledButton, FilledTonalButton, OutlinedButton, TextButton,
// ElevatedButton, or IconButton. By default it is styled by the active
// Material 3 ColorScheme; the chainable setters override that per button.
type ButtonWidget struct {
	handler      func(Event)
	Label        string
	icon         string
	fontSize     float64
	borderRadius float64
	bgColor      style.Color
	variant      fugov1.ButtonVariant
	enabled      bool
	bgColorSet   bool
	baseWidget
}

func newButton(label string, variant fugov1.ButtonVariant) *ButtonWidget {
	return &ButtonWidget{Label: label, variant: variant, enabled: true}
}

// Button creates a Material 3 filled button — an alias of FilledButton.
func Button(label string) *ButtonWidget {
	return newButton(label, fugov1.ButtonVariant_BUTTON_FILLED)
}

// FilledButton creates a high-emphasis Material 3 filled button.
func FilledButton(label string) *ButtonWidget {
	return newButton(label, fugov1.ButtonVariant_BUTTON_FILLED)
}

// FilledTonalButton creates a medium-emphasis Material 3 tonal button.
func FilledTonalButton(label string) *ButtonWidget {
	return newButton(label, fugov1.ButtonVariant_BUTTON_FILLED_TONAL)
}

// OutlinedButton creates a Material 3 outlined button.
func OutlinedButton(label string) *ButtonWidget {
	return newButton(label, fugov1.ButtonVariant_BUTTON_OUTLINED)
}

// TextButton creates a low-emphasis Material 3 text button.
func TextButton(label string) *ButtonWidget {
	return newButton(label, fugov1.ButtonVariant_BUTTON_TEXT)
}

// ElevatedButton creates a Material 3 elevated button.
func ElevatedButton(label string) *ButtonWidget {
	return newButton(label, fugov1.ButtonVariant_BUTTON_ELEVATED)
}

// IconButton creates an icon-only Material button. icon is an icon name
// (see the names recognized by Icon).
func IconButton(icon string) *ButtonWidget {
	b := newButton("", fugov1.ButtonVariant_BUTTON_ICON)
	b.icon = icon

	return b
}

// OnClick registers the handler invoked when the button is tapped and returns the widget for chaining.
func (b *ButtonWidget) OnClick(handler func(Event)) *ButtonWidget {
	b.handler = handler

	return b
}

// Icon sets a leading icon by name (for icon+label buttons) and returns the widget for chaining.
func (b *ButtonWidget) Icon(name string) *ButtonWidget {
	b.icon = name

	return b
}

// FontSize sets the label font size in logical pixels and returns the widget for chaining.
func (b *ButtonWidget) FontSize(v float64) *ButtonWidget {
	b.fontSize = v

	return b
}

// BgColor overrides the background color (otherwise the M3 ColorScheme decides) and returns the widget for chaining.
func (b *ButtonWidget) BgColor(c style.Color) *ButtonWidget {
	b.bgColor = c
	b.bgColorSet = true

	return b
}

// BorderRadius sets the corner radius in logical pixels and returns the widget for chaining.
func (b *ButtonWidget) BorderRadius(v float64) *ButtonWidget {
	b.borderRadius = v

	return b
}

// Enabled toggles whether the button is interactive and returns the widget for chaining.
func (b *ButtonWidget) Enabled(v bool) *ButtonWidget {
	b.enabled = v

	return b
}

func (b *ButtonWidget) isWidget() {}

func (b *ButtonWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnClick handler has been registered.
func (b *ButtonWidget) HasHandler() bool { return b.handler != nil }

// Handle dispatches event to the registered OnClick handler, if any.
func (b *ButtonWidget) Handle(event Event) {
	if b.handler != nil {
		b.handler(event)
	}
}

func (b *ButtonWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	b.id = *counter

	bgColor := ""
	if b.bgColorSet {
		bgColor = b.bgColor.String()
	}

	props, _ := proto.Marshal(&fugov1.ButtonProps{
		Label:        b.Label,
		BgColor:      bgColor,
		FontSize:     b.fontSize,
		BorderRadius: b.borderRadius,
		Variant:      b.variant,
		Icon:         b.icon,
		Enabled:      b.enabled,
	})

	return []*fugov1.WidgetNode{{
		Id:    b.id,
		Key:   b.key,
		Type:  fugov1.WidgetType_BUTTON,
		Props: props,
	}}
}
