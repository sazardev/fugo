package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// TextFieldWidget is an editable single-line text input. Build one with TextField.
type TextFieldWidget struct {
	handler     func(Event)
	Value       string
	Placeholder string
	fontSize    float64
	obscure     bool
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
		Value:       t.Value,
		Placeholder: t.Placeholder,
		FontSize:    t.fontSize,
		Obscure:     t.obscure,
	})

	return []*fugov1.WidgetNode{{
		Id:    t.id,
		Key:   t.key,
		Type:  fugov1.WidgetType_TEXTFIELD,
		Props: props,
	}}
}
