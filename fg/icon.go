package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type IconWidget struct {
	Name   string
	size   float64
	color_ style.Color
	baseWidget
}

func Icon(name string) *IconWidget {
	return &IconWidget{
		Name:   name,
		size:   24,
		color_: active.Colors.OnSurface,
	}
}

func (i *IconWidget) Size(v float64) *IconWidget {
	i.size = v

	return i
}

func (i *IconWidget) Color(c style.Color) *IconWidget {
	i.color_ = c

	return i
}

func (i *IconWidget) isWidget()                {}
func (i *IconWidget) widgetChildren() []Widget { return nil }

func (i *IconWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	i.id = *counter

	props, _ := proto.Marshal(&fugov1.IconProps{
		Name:  i.Name,
		Size:  i.size,
		Color: i.color_.String(),
	})

	return []*fugov1.WidgetNode{{
		Id:    i.id,
		Key:   i.key,
		Type:  fugov1.WidgetType_ICON,
		Props: props,
	}}
}
