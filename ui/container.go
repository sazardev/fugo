package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Container struct {
	child        Widget
	BgColor      string
	Padding      float64
	BorderRadius float64
	baseWidget
}

func NewContainer(child Widget) *Container {
	return &Container{BgColor: "#121212", child: child}
}

func (c *Container) isWidget() {}

func (c *Container) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *Container) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	var childIDs []uint32

	var allNodes []*fugov1.WidgetNode

	for _, child := range c.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.ContainerProps{
		BgColor:      c.BgColor,
		Padding:      c.Padding,
		BorderRadius: c.BorderRadius,
	})

	self := &fugov1.WidgetNode{
		Id:       c.id,
		Type:     fugov1.WidgetType_CONTAINER,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
