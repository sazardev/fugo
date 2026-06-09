package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// DropdownWidget is a select control over a list of string items. Build one
// with Dropdown.
type DropdownWidget struct {
	handler func(Event)
	Value   string
	Items   []string
	baseWidget
}

// Dropdown creates a select control offering items.
func Dropdown(items []string) *DropdownWidget {
	return &DropdownWidget{Items: items}
}

// SetValue sets the currently selected item and returns the widget for chaining.
func (d *DropdownWidget) SetValue(v string) *DropdownWidget {
	d.Value = v

	return d
}

// OnChange registers the handler invoked when the selection changes and returns
// the widget for chaining.
func (d *DropdownWidget) OnChange(handler func(Event)) *DropdownWidget {
	d.handler = handler

	return d
}

func (d *DropdownWidget) isWidget()                {}
func (d *DropdownWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (d *DropdownWidget) HasHandler() bool { return d.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (d *DropdownWidget) Handle(event Event) {
	if d.handler != nil {
		d.handler(event)
	}
}

func (d *DropdownWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	props, _ := proto.Marshal(&fugov1.DropdownProps{
		Items: d.Items,
		Value: d.Value,
	})

	return []*fugov1.WidgetNode{{
		Id:    d.id,
		Key:   d.key,
		Type:  fugov1.WidgetType_DROPDOWN,
		Props: props,
	}}
}
