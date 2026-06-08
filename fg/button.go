package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type ButtonWidget struct {
	handler      func(Event)
	Label        string
	bgColor      style.Color
	fontSize     float64
	borderRadius float64
	baseWidget
}

func Button(label string) *ButtonWidget {
	return &ButtonWidget{
		Label:        label,
		bgColor:      active.Colors.Primary,
		fontSize:     active.Typography.Body,
		borderRadius: active.Radius.MD,
	}
}

func (b *ButtonWidget) OnClick(handler func(Event)) *ButtonWidget {
	b.handler = handler

	return b
}

func (b *ButtonWidget) FontSize(v float64) *ButtonWidget {
	b.fontSize = v

	return b
}

func (b *ButtonWidget) BgColor(c style.Color) *ButtonWidget {
	b.bgColor = c

	return b
}

func (b *ButtonWidget) BorderRadius(v float64) *ButtonWidget {
	b.borderRadius = v

	return b
}

func (b *ButtonWidget) isWidget() {}

func (b *ButtonWidget) widgetChildren() []Widget { return nil }

func (b *ButtonWidget) HasHandler() bool { return b.handler != nil }

func (b *ButtonWidget) Handle(event Event) {
	if b.handler != nil {
		b.handler(event)
	}
}

func (b *ButtonWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	b.id = *counter

	props, _ := proto.Marshal(&fugov1.ButtonProps{
		Label:        b.Label,
		BgColor:      b.bgColor.String(),
		FontSize:     b.fontSize,
		BorderRadius: b.borderRadius,
	})

	return []*fugov1.WidgetNode{{
		Id:    b.id,
		Key:   b.key,
		Type:  fugov1.WidgetType_BUTTON,
		Props: props,
	}}
}
