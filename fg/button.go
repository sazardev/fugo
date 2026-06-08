package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ButtonWidget is a tappable button with a text label. Build one with Button.
type ButtonWidget struct {
	handler      func(Event)
	Label        string
	bgColor      style.Color
	fontSize     float64
	borderRadius float64
	baseWidget
}

// Button creates a button with the given label, styled from the active Theme.
func Button(label string) *ButtonWidget {
	return &ButtonWidget{
		Label:        label,
		bgColor:      active.Colors.Primary,
		fontSize:     active.Typography.Body,
		borderRadius: active.Radius.MD,
	}
}

// OnClick registers the handler invoked when the button is tapped and returns the widget for chaining.
func (b *ButtonWidget) OnClick(handler func(Event)) *ButtonWidget {
	b.handler = handler

	return b
}

// FontSize sets the label font size in logical pixels and returns the widget for chaining.
func (b *ButtonWidget) FontSize(v float64) *ButtonWidget {
	b.fontSize = v

	return b
}

// BgColor sets the button background color and returns the widget for chaining.
func (b *ButtonWidget) BgColor(c style.Color) *ButtonWidget {
	b.bgColor = c

	return b
}

// BorderRadius sets the corner radius in logical pixels and returns the widget for chaining.
func (b *ButtonWidget) BorderRadius(v float64) *ButtonWidget {
	b.borderRadius = v

	return b
}

func (b *ButtonWidget) isWidget() {}

func (b *ButtonWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnClick handler has been registered.
func (b *ButtonWidget) HasHandler() bool { return b.handler != nil }

// Handle dispatches event to the registered OnClick handler, if any.
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
