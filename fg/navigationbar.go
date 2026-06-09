package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// NavigationBarWidget is a Material 3 bottom navigation bar. Build one with
// NavigationBar, add destinations with Item, set the selected index, and read
// taps with OnChange (the event data is the selected index). Pass it to
// Scaffold().BottomBar(...).
type NavigationBarWidget struct {
	handler  func(Event)
	icons    []string
	labels   []string
	selected int
	baseWidget
}

// NavigationBar creates an empty bottom navigation bar; add destinations with Item.
func NavigationBar() *NavigationBarWidget {
	return &NavigationBarWidget{}
}

// Item appends a destination with the given icon (see fg.Icons) and label, and returns the widget for chaining.
func (n *NavigationBarWidget) Item(icon, label string) *NavigationBarWidget {
	n.icons = append(n.icons, icon)
	n.labels = append(n.labels, label)

	return n
}

// Selected sets the highlighted destination index and returns the widget for chaining.
func (n *NavigationBarWidget) Selected(index int) *NavigationBarWidget {
	n.selected = index

	return n
}

// OnChange registers the handler invoked when a destination is tapped; the
// event data is the selected index. Returns the widget for chaining.
func (n *NavigationBarWidget) OnChange(handler func(Event)) *NavigationBarWidget {
	n.handler = handler

	return n
}

func (n *NavigationBarWidget) isWidget() {}

func (n *NavigationBarWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler is registered.
func (n *NavigationBarWidget) HasHandler() bool { return n.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (n *NavigationBarWidget) Handle(event Event) {
	if n.handler != nil {
		n.handler(event)
	}
}

func (n *NavigationBarWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	n.id = *counter

	props, _ := proto.Marshal(&fugov1.NavigationBarProps{
		Icons:         n.icons,
		Labels:        n.labels,
		SelectedIndex: int32(n.selected), //nolint:gosec // a small UI destination index
	})

	return []*fugov1.WidgetNode{{
		Id:    n.id,
		Key:   n.key,
		Type:  fugov1.WidgetType_NAVIGATIONBAR,
		Props: props,
	}}
}
