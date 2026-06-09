package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// CheckboxWidget is a labeled checkbox with a boolean checked state. Build one with Checkbox.
type CheckboxWidget struct {
	handler func(Event)
	Label   string
	Checked bool
	baseWidget
}

// Checkbox creates a checkbox with the given label.
func Checkbox(label string) *CheckboxWidget {
	return &CheckboxWidget{Label: label}
}

// OnChange registers the handler invoked when the checked state toggles and returns the widget for chaining.
func (c *CheckboxWidget) OnChange(handler func(Event)) *CheckboxWidget {
	c.handler = handler

	return c
}

// SetChecked sets the checked state and returns the widget for chaining.
func (c *CheckboxWidget) SetChecked(v bool) *CheckboxWidget {
	c.Checked = v

	return c
}

// SetLabel sets the label text and returns the widget for chaining.
func (c *CheckboxWidget) SetLabel(v string) *CheckboxWidget {
	c.Label = v

	return c
}

func (c *CheckboxWidget) isWidget()                {}
func (c *CheckboxWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (c *CheckboxWidget) HasHandler() bool { return c.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (c *CheckboxWidget) Handle(event Event) {
	if c.handler != nil {
		c.handler(event)
	}
}

func (c *CheckboxWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	props, _ := proto.Marshal(&fugov1.CheckboxProps{
		Checked: c.Checked,
		Label:   c.Label,
	})

	return []*fugov1.WidgetNode{{
		Id:    c.id,
		Key:   c.key,
		Type:  fugov1.WidgetType_CHECKBOX,
		Props: props,
	}}
}
