package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Text struct {
	Value string
	Style style.TextStyle
	baseWidget
}

func Text(value string) *Text {
	return &Text{
		Value: value,
		Style: style.NewTextStyle(defaultFontSize, style.Hex("#FFFFFF")),
	}
}

func (t *Text) SetText(value string) {
	t.Value = value
}

func (t *Text) FontSize(v float64) *Text {
	t.Style.FontSize = v

	return t
}

func (t *Text) Color(c style.Color) *Text {
	t.Style.Color = c

	return t
}

func (t *Text) Weight(w style.FontWeight) *Text {
	t.Style.Weight = w

	return t
}

func (t *Text) Align(a style.TextAlign) *Text {
	t.Style.Align = a

	return t
}

func (t *Text) isWidget() {}

func (t *Text) widgetChildren() []Widget { return nil }

func (t *Text) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	props, _ := proto.Marshal(&fugov1.TextProps{
		Value:    t.Value,
		FontSize: t.Style.FontSize,
		Color:    t.Style.Color.String(),
	})

	return []*fugov1.WidgetNode{{
		Id:    t.id,
		Key:   t.key,
		Type:  fugov1.WidgetType_TEXT,
		Props: props,
	}}
}
