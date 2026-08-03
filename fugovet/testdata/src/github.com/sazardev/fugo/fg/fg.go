// Package fg is a minimal stand-in for github.com/sazardev/fugo/fg, just
// enough to exercise fugovet's checks without pulling in the real widget
// tree/protobuf machinery.
package fg

// Event is a minimal stand-in for fg.Event.
type Event struct{ Data []byte }

// Widget is a minimal stand-in for fg.Widget.
type Widget interface {
	isWidget()
}

// TextWidget is a minimal stand-in for fg.TextWidget.
type TextWidget struct{ text string }

// Text is a minimal stand-in for fg.Text.
func Text(s string) *TextWidget { return &TextWidget{text: s} }

// SetText is a minimal stand-in for (*fg.TextWidget).SetText.
func (t *TextWidget) SetText(s string) *TextWidget {
	t.text = s

	return t
}

func (t *TextWidget) isWidget() {}

// ButtonWidget is a minimal stand-in for fg.ButtonWidget.
type ButtonWidget struct {
	handler func(Event)
}

func newButton() *ButtonWidget { return &ButtonWidget{} }

// Button is a minimal stand-in for fg.Button.
func Button(label string) *ButtonWidget { return newButton() }

// FilledButton is a minimal stand-in for fg.FilledButton.
func FilledButton(label string) *ButtonWidget { return newButton() }

// FilledTonalButton is a minimal stand-in for fg.FilledTonalButton.
func FilledTonalButton(label string) *ButtonWidget { return newButton() }

// OutlinedButton is a minimal stand-in for fg.OutlinedButton.
func OutlinedButton(label string) *ButtonWidget { return newButton() }

// TextButton is a minimal stand-in for fg.TextButton.
func TextButton(label string) *ButtonWidget { return newButton() }

// ElevatedButton is a minimal stand-in for fg.ElevatedButton.
func ElevatedButton(label string) *ButtonWidget { return newButton() }

// IconButton is a minimal stand-in for fg.IconButton.
func IconButton(icon string) *ButtonWidget { return newButton() }

// OnClick is a minimal stand-in for (*fg.ButtonWidget).OnClick.
func (b *ButtonWidget) OnClick(h func(Event)) *ButtonWidget {
	b.handler = h

	return b
}

// BgColor is a minimal stand-in for (*fg.ButtonWidget).BgColor.
func (b *ButtonWidget) BgColor(c string) *ButtonWidget { return b }

func (b *ButtonWidget) isWidget() {}

// Row is a minimal stand-in for fg.Row.
func Row(children ...Widget) Widget { return nil }
