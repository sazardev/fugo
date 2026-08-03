package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// KeyboardType selects which on-screen/soft keyboard layout and input
// validation hint a TextField requests. Values match
// TextFieldProps.keyboard_type on the wire.
type KeyboardType int32

// KeyboardType values, matching TextFieldProps.keyboard_type (0-4).
const (
	KeyboardText KeyboardType = iota
	KeyboardNumber
	KeyboardEmail
	KeyboardMultiline
	KeyboardPhone
)

// TextFieldWidget is an editable text input, single-line by default. Build one with TextField.
type TextFieldWidget struct {
	handler      func(Event)
	Value        string
	Placeholder  string
	fontSize     float64
	obscure      bool
	maxLines     int
	prefixIcon   string
	suffixIcon   string
	errorText    string
	keyboardType KeyboardType
	baseWidget
}

// TextField creates a text input showing placeholder when empty, styled from the active Theme.
func TextField(placeholder string) *TextFieldWidget {
	return &TextFieldWidget{
		Placeholder: placeholder,
		fontSize:    active.Typography.Body,
	}
}

// OnChange registers the handler invoked when the text changes and returns the widget for chaining.
func (t *TextFieldWidget) OnChange(handler func(Event)) *TextFieldWidget {
	t.handler = handler

	return t
}

// FontSize sets the input font size in logical pixels and returns the widget for chaining.
func (t *TextFieldWidget) FontSize(v float64) *TextFieldWidget {
	t.fontSize = v

	return t
}

// SetValue sets the current text and returns the widget for chaining.
func (t *TextFieldWidget) SetValue(v string) *TextFieldWidget {
	t.Value = v

	return t
}

// Obscure toggles password-style masking of the input and returns the widget for chaining.
func (t *TextFieldWidget) Obscure(v bool) *TextFieldWidget {
	t.obscure = v

	return t
}

// MaxLines sets the number of visible lines; n <= 1 (the default) is a
// single-line field, n > 1 makes it multiline. Returns the widget for chaining.
func (t *TextFieldWidget) MaxLines(n int) *TextFieldWidget {
	t.maxLines = n

	return t
}

// PrefixIcon sets a leading icon (by name, e.g. fg.Icons.Search); empty
// clears it. Returns the widget for chaining.
func (t *TextFieldWidget) PrefixIcon(name string) *TextFieldWidget {
	t.prefixIcon = name

	return t
}

// SuffixIcon sets a trailing icon (by name); empty clears it. Returns the
// widget for chaining.
func (t *TextFieldWidget) SuffixIcon(name string) *TextFieldWidget {
	t.suffixIcon = name

	return t
}

// KeyboardType sets the requested keyboard layout/input hint and returns the
// widget for chaining.
func (t *TextFieldWidget) KeyboardType(v KeyboardType) *TextFieldWidget {
	t.keyboardType = v

	return t
}

// SetError sets the validation error shown below the field (owned entirely by
// Go — call it from an OnChange handler); empty clears it. Returns the widget
// for chaining.
func (t *TextFieldWidget) SetError(msg string) *TextFieldWidget {
	t.errorText = msg

	return t
}

func (t *TextFieldWidget) isWidget()                {}
func (t *TextFieldWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (t *TextFieldWidget) HasHandler() bool { return t.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (t *TextFieldWidget) Handle(event Event) {
	if t.handler != nil {
		t.handler(event)
	}
}

func (t *TextFieldWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	props, _ := proto.Marshal(&fugov1.TextFieldProps{
		Value:        t.Value,
		Placeholder:  t.Placeholder,
		FontSize:     t.fontSize,
		Obscure:      t.obscure,
		MaxLines:     int32(t.maxLines), //nolint:gosec // small line count
		PrefixIcon:   t.prefixIcon,
		SuffixIcon:   t.suffixIcon,
		ErrorText:    t.errorText,
		KeyboardType: int32(t.keyboardType),
	})

	return []*fugov1.WidgetNode{{
		Id:    t.id,
		Key:   t.key,
		Type:  fugov1.WidgetType_TEXTFIELD,
		Props: props,
	}}
}
