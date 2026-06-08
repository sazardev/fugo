package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Text struct {
	Value    string
	Color    string
	FontSize float64
	baseWidget
}

func NewText(value string) *Text {
	return &Text{Value: value, FontSize: 14, Color: "#FFFFFF"}
}

func (t *Text) isWidget() {}

func (t *Text) widgetChildren() []Widget { return nil }

func (t *Text) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	props, _ := proto.Marshal(&fugov1.TextProps{
		Value:    t.Value,
		FontSize: t.FontSize,
		Color:    t.Color,
	})

	return []*fugov1.WidgetNode{{
		Id:    t.id,
		Type:  fugov1.WidgetType_TEXT,
		Props: props,
	}}
}

func (t *Text) SetText(value string) {
	t.Value = value
}
