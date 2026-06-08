package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type TextFieldWidget struct {
	handler     func(Event)
	Value       string
	Placeholder string
	fontSize    float64
	obscure     bool
	baseWidget
}

func TextField(placeholder string) *TextFieldWidget {
	return &TextFieldWidget{
		Placeholder: placeholder,
		fontSize:    active.Typography.Body,
	}
}

func (t *TextFieldWidget) OnChange(handler func(Event)) *TextFieldWidget {
	t.handler = handler

	return t
}

func (t *TextFieldWidget) FontSize(v float64) *TextFieldWidget {
	t.fontSize = v

	return t
}

func (t *TextFieldWidget) SetValue(v string) *TextFieldWidget {
	t.Value = v

	return t
}

func (t *TextFieldWidget) Obscure(v bool) *TextFieldWidget {
	t.obscure = v

	return t
}

func (t *TextFieldWidget) isWidget()                {}
func (t *TextFieldWidget) widgetChildren() []Widget { return nil }
func (t *TextFieldWidget) HasHandler() bool         { return t.handler != nil }

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
