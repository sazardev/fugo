package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type ImageWidget struct {
	Src    string
	width  float64
	height float64
	baseWidget
}

func Image(src string) *ImageWidget {
	return &ImageWidget{Src: src}
}

func (i *ImageWidget) Size(w, h float64) *ImageWidget {
	i.width = w
	i.height = h

	return i
}

func (i *ImageWidget) Width(v float64) *ImageWidget {
	i.width = v

	return i
}

func (i *ImageWidget) Height(v float64) *ImageWidget {
	i.height = v

	return i
}

func (i *ImageWidget) isWidget()                {}
func (i *ImageWidget) widgetChildren() []Widget { return nil }

func (i *ImageWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	i.id = *counter

	props, _ := proto.Marshal(&fugov1.ImageProps{
		Src:    i.Src,
		Width:  i.width,
		Height: i.height,
	})

	return []*fugov1.WidgetNode{{
		Id:    i.id,
		Key:   i.key,
		Type:  fugov1.WidgetType_IMAGE,
		Props: props,
	}}
}
