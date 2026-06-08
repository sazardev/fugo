package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Image struct {
	Src    string
	Width  float64
	Height float64
	baseWidget
}

func NewImage(src string) *Image {
	return &Image{Src: src}
}

func (i *Image) WithSize(w, h float64) *Image {
	i.Width = w
	i.Height = h

	return i
}

func (i *Image) WithWidth(v float64) *Image {
	i.Width = v

	return i
}

func (i *Image) WithHeight(v float64) *Image {
	i.Height = v

	return i
}

func (i *Image) isWidget()                {}
func (i *Image) widgetChildren() []Widget { return nil }

func (i *Image) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	i.id = *counter

	props, _ := proto.Marshal(&fugov1.ImageProps{
		Src:    i.Src,
		Width:  i.Width,
		Height: i.Height,
	})

	return []*fugov1.WidgetNode{{
		Id:    i.id,
		Key:   i.key,
		Type:  fugov1.WidgetType_IMAGE,
		Props: props,
	}}
}
