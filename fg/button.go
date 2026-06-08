package ui

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

const (
	defaultFontSize     = 14.0
	defaultBorderRadius = 8.0
)

type Button struct {
	handler      func(Event)
	Label        string
	BgColor      style.Color
	FontSize     float64
	BorderRadius float64
	baseWidget
}

func NewButton(label string) *Button {
	return &Button{
		Label:        label,
		BgColor:      style.Hex("#3B82F6"),
		FontSize:     defaultFontSize,
		BorderRadius: defaultBorderRadius,
	}
}

func (b *Button) OnClick(handler func(Event)) *Button {
	b.handler = handler

	return b
}

func (b *Button) WithFontSize(v float64) *Button {
	b.FontSize = v

	return b
}

func (b *Button) WithBgColor(c style.Color) *Button {
	b.BgColor = c

	return b
}

func (b *Button) WithBorderRadius(v float64) *Button {
	b.BorderRadius = v

	return b
}

func (b *Button) isWidget() {}

func (b *Button) widgetChildren() []Widget { return nil }

func (b *Button) HasHandler() bool { return b.handler != nil }

func (b *Button) Handle(event Event) {
	if b.handler != nil {
		b.handler(event)
	}
}

func (b *Button) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	b.id = *counter

	props, _ := proto.Marshal(&fugov1.ButtonProps{
		Label:        b.Label,
		BgColor:      b.BgColor.String(),
		FontSize:     b.FontSize,
		BorderRadius: b.BorderRadius,
	})

	return []*fugov1.WidgetNode{{
		Id:    b.id,
		Key:   b.key,
		Type:  fugov1.WidgetType_BUTTON,
		Props: props,
	}}
}
