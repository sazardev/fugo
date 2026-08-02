package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// TextOverflow selects how a Text widget handles content that doesn't fit
// its maxLines. Values match TextProps.overflow on the wire.
type TextOverflow int32

// TextOverflow values, matching TextProps.overflow (0-3).
const (
	OverflowClip TextOverflow = iota
	OverflowEllipsis
	OverflowFade
	OverflowVisible
)

// TextDecoration selects a line decoration for a Text widget. Values match
// TextProps.decoration on the wire.
type TextDecoration int32

// TextDecoration values, matching TextProps.decoration (0-3).
const (
	DecorationNone TextDecoration = iota
	DecorationUnderline
	DecorationLineThrough
	DecorationOverline
)

// TextWidget is a run of styled text. Build one with Text.
type TextWidget struct {
	Value string
	Style style.TextStyle
	// colorSet records whether Color was set explicitly. When false the color
	// is left unset on the wire so the Material 3 text theme styles it (so text
	// reads correctly in both light and dark mode).
	colorSet      bool
	maxLines      int
	overflow      TextOverflow
	decoration    TextDecoration
	letterSpacing float64
	italic        bool
	baseWidget
}

// Text creates a text widget showing value. Its color is left to the active
// Material 3 theme unless overridden with Color.
func Text(value string) *TextWidget {
	return &TextWidget{
		Value: value,
		Style: style.NewTextStyle(active.Typography.Body, style.Color{}),
	}
}

// SetText replaces the displayed text; call ctx.Update to render the change.
func (t *TextWidget) SetText(value string) {
	t.Value = value
}

// FontSize sets the font size in logical pixels and returns the widget for chaining.
func (t *TextWidget) FontSize(v float64) *TextWidget {
	t.Style.FontSize = v

	return t
}

// Color sets the text color and returns the widget for chaining.
func (t *TextWidget) Color(c style.Color) *TextWidget {
	t.Style.Color = c
	t.colorSet = true

	return t
}

// Weight sets the font weight and returns the widget for chaining.
func (t *TextWidget) Weight(w style.FontWeight) *TextWidget {
	t.Style.Weight = w

	return t
}

// Align sets the horizontal text alignment and returns the widget for chaining.
func (t *TextWidget) Align(a style.TextAlign) *TextWidget {
	t.Style.Align = a

	return t
}

// MaxLines caps the number of lines rendered before overflow kicks in; 0
// (the default) means unlimited. Returns the widget for chaining.
func (t *TextWidget) MaxLines(n int) *TextWidget {
	t.maxLines = n

	return t
}

// Overflow sets how content past MaxLines is handled and returns the widget
// for chaining.
func (t *TextWidget) Overflow(v TextOverflow) *TextWidget {
	t.overflow = v

	return t
}

// Decoration sets a line decoration (underline, line-through, overline) and
// returns the widget for chaining.
func (t *TextWidget) Decoration(v TextDecoration) *TextWidget {
	t.decoration = v

	return t
}

// LetterSpacing sets extra spacing between characters in logical pixels; 0
// is the default. Returns the widget for chaining.
func (t *TextWidget) LetterSpacing(v float64) *TextWidget {
	t.letterSpacing = v

	return t
}

// Italic toggles italic styling and returns the widget for chaining.
func (t *TextWidget) Italic(v bool) *TextWidget {
	t.italic = v

	return t
}

func (t *TextWidget) isWidget() {}

func (t *TextWidget) widgetChildren() []Widget { return nil }

func (t *TextWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	color := ""
	if t.colorSet {
		color = t.Style.Color.String()
	}

	props, _ := proto.Marshal(&fugov1.TextProps{
		Value:         t.Value,
		FontSize:      t.Style.FontSize,
		Color:         color,
		FontWeight:    int32(t.Style.Weight), //nolint:gosec // bounded weight (100..900)
		TextAlign:     int32(t.Style.Align),  //nolint:gosec // bounded align (0..2)
		MaxLines:      int32(t.maxLines),     //nolint:gosec // small line count
		Overflow:      int32(t.overflow),
		Decoration:    int32(t.decoration),
		LetterSpacing: t.letterSpacing,
		Italic:        t.italic,
	})

	return []*fugov1.WidgetNode{{
		Id:    t.id,
		Key:   t.key,
		Type:  fugov1.WidgetType_TEXT,
		Props: props,
	}}
}
