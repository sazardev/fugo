package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type CheckboxWidget struct {
	handler func(Event)
	Label   string
	Checked bool
	baseWidget
}

func Checkbox(label string) *CheckboxWidget {
	return &CheckboxWidget{Label: label}
}

func (c *CheckboxWidget) OnChange(handler func(Event)) *CheckboxWidget {
	c.handler = handler

	return c
}

func (c *CheckboxWidget) SetChecked(v bool) *CheckboxWidget {
	c.Checked = v

	return c
}

func (c *CheckboxWidget) SetLabel(v string) *CheckboxWidget {
	c.Label = v

	return c
}

func (c *CheckboxWidget) isWidget()                {}
func (c *CheckboxWidget) widgetChildren() []Widget { return nil }
func (c *CheckboxWidget) HasHandler() bool         { return c.handler != nil }

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
