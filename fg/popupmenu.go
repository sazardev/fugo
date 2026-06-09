package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// PopupMenuWidget is a Material overflow/context menu opened from an icon. Build
// one with PopupMenuButton, add entries with Item, and read the chosen entry
// with OnSelected (the event data is the selected value).
type PopupMenuWidget struct {
	handler func(Event)
	icon    string
	values  []string
	labels  []string
	baseWidget
}

// PopupMenuButton creates a popup menu opened from the named icon (defaults to a
// three-dot "more" icon when empty).
func PopupMenuButton(icon string) *PopupMenuWidget {
	return &PopupMenuWidget{icon: icon}
}

// Item appends a menu entry with the given value and label, and returns the widget for chaining.
func (p *PopupMenuWidget) Item(value, label string) *PopupMenuWidget {
	p.values = append(p.values, value)
	p.labels = append(p.labels, label)

	return p
}

// OnSelected registers the handler invoked when an entry is chosen; the event
// data is the selected value. Returns the widget for chaining.
func (p *PopupMenuWidget) OnSelected(handler func(Event)) *PopupMenuWidget {
	p.handler = handler

	return p
}

func (p *PopupMenuWidget) isWidget()                {}
func (p *PopupMenuWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnSelected handler is registered.
func (p *PopupMenuWidget) HasHandler() bool { return p.handler != nil }

// Handle dispatches event to the registered OnSelected handler, if any.
func (p *PopupMenuWidget) Handle(event Event) {
	if p.handler != nil {
		p.handler(event)
	}
}

func (p *PopupMenuWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	p.id = *counter

	props, _ := proto.Marshal(&fugov1.PopupMenuProps{
		Icon:   p.icon,
		Values: p.values,
		Labels: p.labels,
	})

	return []*fugov1.WidgetNode{{
		Id:    p.id,
		Key:   p.key,
		Type:  fugov1.WidgetType_POPUPMENU,
		Props: props,
	}}
}
