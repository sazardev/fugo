package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ChipWidget is a Material 3 chip. Build one with Chip. Marking it Selected
// renders a FilterChip; the tap/delete affordances forward to the OnTap handler
// (distinguish them via Event.EventType: "onTap" vs "onDeleted").
type ChipWidget struct {
	handler   func(Event)
	Label     string
	selected  bool
	deletable bool
	baseWidget
}

// Chip creates a chip with the given label.
func Chip(label string) *ChipWidget {
	return &ChipWidget{Label: label}
}

// Selected marks the chip selected (rendering a FilterChip) and returns the widget for chaining.
func (c *ChipWidget) Selected(v bool) *ChipWidget {
	c.selected = v

	return c
}

// Deletable shows a delete affordance and returns the widget for chaining.
func (c *ChipWidget) Deletable(v bool) *ChipWidget {
	c.deletable = v

	return c
}

// OnTap registers the handler for taps and deletes and returns the widget for chaining.
func (c *ChipWidget) OnTap(handler func(Event)) *ChipWidget {
	c.handler = handler

	return c
}

func (c *ChipWidget) isWidget() {}

func (c *ChipWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnTap handler has been registered.
func (c *ChipWidget) HasHandler() bool { return c.handler != nil }

// Handle dispatches event to the registered OnTap handler, if any.
func (c *ChipWidget) Handle(event Event) {
	if c.handler != nil {
		c.handler(event)
	}
}

func (c *ChipWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	props, _ := proto.Marshal(&fugov1.ChipProps{
		Label:     c.Label,
		Selected:  c.selected,
		Deletable: c.deletable,
	})

	return []*fugov1.WidgetNode{{
		Id:    c.id,
		Key:   c.key,
		Type:  fugov1.WidgetType_CHIP,
		Props: props,
	}}
}
