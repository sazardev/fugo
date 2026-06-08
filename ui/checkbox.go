package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Checkbox struct {
	handler func(Event)
	Label   string
	Checked bool
	baseWidget
}

func NewCheckbox(label string) *Checkbox {
	return &Checkbox{Label: label}
}

func (c *Checkbox) OnChange(handler func(Event)) *Checkbox {
	c.handler = handler

	return c
}

func (c *Checkbox) WithChecked(v bool) *Checkbox {
	c.Checked = v

	return c
}

func (c *Checkbox) WithLabel(v string) *Checkbox {
	c.Label = v

	return c
}

func (c *Checkbox) isWidget()                {}
func (c *Checkbox) widgetChildren() []Widget { return nil }
func (c *Checkbox) HasHandler() bool         { return c.handler != nil }

func (c *Checkbox) Handle(event Event) {
	if c.handler != nil {
		c.handler(event)
	}
}

func (c *Checkbox) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
