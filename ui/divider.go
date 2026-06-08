package ui

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Divider struct {
	Thickness float64
	Color     style.Color
	baseWidget
}

func NewDivider() *Divider {
	return &Divider{
		Thickness: 1,
	}
}

func (d *Divider) WithThickness(v float64) *Divider {
	d.Thickness = v

	return d
}

func (d *Divider) WithColor(c style.Color) *Divider {
	d.Color = c

	return d
}

func (d *Divider) isWidget()                {}
func (d *Divider) widgetChildren() []Widget { return nil }

func (d *Divider) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	props, _ := proto.Marshal(&fugov1.DividerProps{
		Thickness: d.Thickness,
		Color:     d.Color.String(),
	})

	return []*fugov1.WidgetNode{{
		Id:    d.id,
		Key:   d.key,
		Type:  fugov1.WidgetType_DIVIDER,
		Props: props,
	}}
}
