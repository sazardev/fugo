package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type ContainerWidget struct {
	child        Widget
	bgColor      style.Color
	Padding      style.EdgeInsets
	borderRadius float64
	baseWidget
}

func Container(child Widget) *ContainerWidget {
	return &ContainerWidget{
		child:   child,
		bgColor: style.Hex("#121212"),
	}
}

func (c *ContainerWidget) BgColor(v style.Color) *ContainerWidget {
	c.bgColor = v

	return c
}

func (c *ContainerWidget) Pad(v style.EdgeInsets) *ContainerWidget {
	c.Padding = v

	return c
}

func (c *ContainerWidget) BorderRadius(v float64) *ContainerWidget {
	c.borderRadius = v

	return c
}

func (c *ContainerWidget) isWidget() {}

func (c *ContainerWidget) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *ContainerWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		BgColor:      c.bgColor.String(),
		Padding:      c.Padding.Top,
		BorderRadius: c.borderRadius,
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
