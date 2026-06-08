package ui

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Icon struct {
	Name  string
	Size  float64
	Color style.Color
	baseWidget
}

func NewIcon(name string) *Icon {
	return &Icon{
		Name: name,
		Size: 24,
	}
}

func (i *Icon) WithSize(v float64) *Icon {
	i.Size = v

	return i
}

func (i *Icon) WithColor(c style.Color) *Icon {
	i.Color = c

	return i
}

func (i *Icon) isWidget()                {}
func (i *Icon) widgetChildren() []Widget { return nil }

func (i *Icon) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	i.id = *counter

	props, _ := proto.Marshal(&fugov1.IconProps{
		Name:  i.Name,
		Size:  i.Size,
		Color: i.Color.String(),
	})

	return []*fugov1.WidgetNode{{
		Id:    i.id,
		Key:   i.key,
		Type:  fugov1.WidgetType_ICON,
		Props: props,
	}}
}
