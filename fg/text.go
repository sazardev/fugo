package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type TextWidget struct {
	Value string
	Style style.TextStyle
	baseWidget
}

func Text(value string) *TextWidget {
	return &TextWidget{
		Value: value,
		Style: style.NewTextStyle(active.Typography.Body, active.Colors.OnSurface),
	}
}

func (t *TextWidget) SetText(value string) {
	t.Value = value
}

func (t *TextWidget) FontSize(v float64) *TextWidget {
	t.Style.FontSize = v

	return t
}

func (t *TextWidget) Color(c style.Color) *TextWidget {
	t.Style.Color = c

	return t
}

func (t *TextWidget) Weight(w style.FontWeight) *TextWidget {
	t.Style.Weight = w

	return t
}

func (t *TextWidget) Align(a style.TextAlign) *TextWidget {
	t.Style.Align = a

	return t
}

func (t *TextWidget) isWidget() {}

func (t *TextWidget) widgetChildren() []Widget { return nil }

func (t *TextWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
