package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type DividerWidget struct {
	thickness float64
	color_    style.Color
	baseWidget
}

func Divider() *DividerWidget {
	return &DividerWidget{
		thickness: 1,
		color_:    active.Colors.Border,
	}
}

func (d *DividerWidget) Thickness(v float64) *DividerWidget {
	d.thickness = v

	return d
}

func (d *DividerWidget) Color(c style.Color) *DividerWidget {
	d.color_ = c

	return d
}

func (d *DividerWidget) isWidget()                {}
func (d *DividerWidget) widgetChildren() []Widget { return nil }

func (d *DividerWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	props, _ := proto.Marshal(&fugov1.DividerProps{
		Thickness: d.thickness,
		Color:     d.color_.String(),
	})

	return []*fugov1.WidgetNode{{
		Id:    d.id,
		Key:   d.key,
		Type:  fugov1.WidgetType_DIVIDER,
		Props: props,
	}}
}
