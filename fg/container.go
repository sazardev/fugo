package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Container struct {
	child        Widget
	BgColor      style.Color
	Padding      style.EdgeInsets
	BorderRadius float64
	baseWidget
}

func Container(child Widget) *Container {
	return &Container{
		child:   child,
		BgColor: style.Hex("#121212"),
	}
}

func (c *Container) BgColor(v style.Color) *Container {
	c.BgColor = v

	return c
}

func (c *Container) Pad(v style.EdgeInsets) *Container {
	c.Padding = v

	return c
}

func (c *Container) BorderRadius(v float64) *Container {
	c.BorderRadius = v

	return c
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
		BgColor:      c.BgColor.String(),
		Padding:      c.Padding.Top,
		BorderRadius: c.BorderRadius,
	})

	self := &fugov1.WidgetNode{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_CONTAINER,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
