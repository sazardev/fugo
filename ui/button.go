package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Button struct {
	handler      func(Event)
	Label        string
	BgColor      string
	FontSize     float64
	BorderRadius float64
	baseWidget
}

func NewButton(label string) *Button {
	return &Button{
		Label:        label,
		BgColor:      "#3B82F6",
		FontSize:     14,
		BorderRadius: 8,
	}
}

func (b *Button) isWidget() {}

func (b *Button) widgetChildren() []Widget { return nil }

func (b *Button) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	b.id = *counter

	props, _ := proto.Marshal(&fugov1.ButtonProps{
		Label:        b.Label,
		BgColor:      b.BgColor,
		FontSize:     b.FontSize,
		BorderRadius: b.BorderRadius,
	})

	return []*fugov1.WidgetNode{{
		Id:    b.id,
		Type:  fugov1.WidgetType_BUTTON,
		Props: props,
	}}
}

func (b *Button) OnClick(handler func(Event)) *Button {
	b.handler = handler

	return b
}

func (b *Button) HasHandler() bool { return b.handler != nil }

func (b *Button) Handle(event Event) {
	if b.handler != nil {
		b.handler(event)
	}
}
