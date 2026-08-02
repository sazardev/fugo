package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// NavigationRailWidget is a Material 3 side navigation rail. Build one with
// NavigationRail, add destinations with Item, set the selected index, and
// read taps with OnChange (the event data is the selected index).
type NavigationRailWidget struct {
	handler  func(Event)
	icons    []string
	labels   []string
	selected int
	extended bool
	baseWidget
}

// NavigationRail creates an empty navigation rail; add destinations with Item.
func NavigationRail() *NavigationRailWidget {
	return &NavigationRailWidget{}
}

// Item appends a destination with the given icon (see fg.Icons) and label, and returns the widget for chaining.
func (n *NavigationRailWidget) Item(icon, label string) *NavigationRailWidget {
	n.icons = append(n.icons, icon)
	n.labels = append(n.labels, label)

	return n
}

// Selected sets the highlighted destination index and returns the widget for chaining.
func (n *NavigationRailWidget) Selected(index int32) *NavigationRailWidget {
	n.selected = int(index)

	return n
}

// Extended sets whether the rail shows its extended (labeled) form and returns the widget for chaining.
func (n *NavigationRailWidget) Extended(v bool) *NavigationRailWidget {
	n.extended = v

	return n
}

// OnChange registers the handler invoked when a destination is tapped; the
// event data is the selected index. Returns the widget for chaining.
func (n *NavigationRailWidget) OnChange(handler func(Event)) *NavigationRailWidget {
	n.handler = handler

	return n
}

func (n *NavigationRailWidget) isWidget() {}

func (n *NavigationRailWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler is registered.
func (n *NavigationRailWidget) HasHandler() bool { return n.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (n *NavigationRailWidget) Handle(event Event) {
	if n.handler != nil {
		n.handler(event)
	}
}

func (n *NavigationRailWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	n.id = *counter

	props, err := proto.Marshal(&fugov1.NavigationRailProps{
		Icons:         n.icons,
		Labels:        n.labels,
		SelectedIndex: int32(n.selected), //nolint:gosec // a small UI destination index
		Extended:      n.extended,
	})
	if err != nil {
		flog.Errorf("marshal NavigationRailProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    n.id,
		Key:   n.key,
		Type:  fugov1.WidgetType_NAVIGATIONRAIL,
		Props: props,
	}}
}
