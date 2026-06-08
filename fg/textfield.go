package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type TextField struct {
	handler     func(Event)
	Value       string
	Placeholder string
	FontSize    float64
	Obscure     bool
	baseWidget
}

func TextField(placeholder string) *TextField {
	return &TextField{
		Placeholder: placeholder,
		FontSize:    defaultFontSize,
	}
}

func (t *TextField) OnChange(handler func(Event)) *TextField {
	t.handler = handler

	return t
}

func (t *TextField) FontSize(v float64) *TextField {
	t.FontSize = v

	return t
}

func (t *TextField) SetValue(v string) *TextField {
	t.Value = v

	return t
}

func (t *TextField) Obscure(v bool) *TextField {
	t.Obscure = v

	return t
}

func (t *TextField) isWidget()                {}
func (t *TextField) widgetChildren() []Widget { return nil }
func (t *TextField) HasHandler() bool         { return t.handler != nil }

func (t *TextField) Handle(event Event) {
	if t.handler != nil {
		t.handler(event)
	}
}

func (t *TextField) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	props, _ := proto.Marshal(&fugov1.TextFieldProps{
		Value:       t.Value,
		Placeholder: t.Placeholder,
		FontSize:    t.FontSize,
		Obscure:     t.Obscure,
	})

	return []*fugov1.WidgetNode{{
		Id:    t.id,
		Key:   t.key,
		Type:  fugov1.WidgetType_TEXTFIELD,
		Props: props,
	}}
}
