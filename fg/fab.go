package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// FabWidget is a Material 3 floating action button. Build one with
// FloatingActionButton. Setting a Label renders an extended FAB.
type FabWidget struct {
	handler func(Event)
	icon    string
	label   string
	mini    bool
	baseWidget
}

// FloatingActionButton creates a FAB showing the named icon (see Icon names).
func FloatingActionButton(icon string) *FabWidget {
	return &FabWidget{icon: icon}
}

// OnClick registers the tap handler and returns the widget for chaining.
func (f *FabWidget) OnClick(handler func(Event)) *FabWidget {
	f.handler = handler

	return f
}

// Label sets a text label, turning the FAB into an extended FAB, and returns the widget for chaining.
func (f *FabWidget) Label(text string) *FabWidget {
	f.label = text

	return f
}

// Mini renders a smaller FAB and returns the widget for chaining.
func (f *FabWidget) Mini(v bool) *FabWidget {
	f.mini = v

	return f
}

func (f *FabWidget) isWidget() {}

func (f *FabWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnClick handler has been registered.
func (f *FabWidget) HasHandler() bool { return f.handler != nil }

// Handle dispatches event to the registered OnClick handler, if any.
func (f *FabWidget) Handle(event Event) {
	if f.handler != nil {
		f.handler(event)
	}
}

func (f *FabWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	f.id = *counter

	props, _ := proto.Marshal(&fugov1.FabProps{
		Icon:  f.icon,
		Label: f.label,
		Mini:  f.mini,
	})

	return []*fugov1.WidgetNode{{
		Id:    f.id,
		Key:   f.key,
		Type:  fugov1.WidgetType_FLOATINGACTIONBUTTON,
		Props: props,
	}}
}
