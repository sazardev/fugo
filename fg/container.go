package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ContainerWidget wraps a single child with a background, padding, and rounded corners. Build one with Container.
type ContainerWidget struct {
	child        Widget
	bgColor      style.Color
	Padding      style.EdgeInsets
	borderRadius float64
	baseWidget
}

// Container creates a container wrapping child.
func Container(child Widget) *ContainerWidget {
	return &ContainerWidget{
		child:   child,
		bgColor: style.Hex("#121212"),
	}
}

// BgColor sets the background color and returns the widget for chaining.
func (c *ContainerWidget) BgColor(v style.Color) *ContainerWidget {
	c.bgColor = v

	return c
}

// Pad sets the inner padding and returns the widget for chaining.
func (c *ContainerWidget) Pad(v style.EdgeInsets) *ContainerWidget {
	c.Padding = v

	return c
}

// BorderRadius sets the corner radius in logical pixels and returns the widget for chaining.
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
		BorderRadius: c.borderRadius,
		PadTop:       c.Padding.Top,
		PadRight:     c.Padding.Right,
		PadBottom:    c.Padding.Bottom,
		PadLeft:      c.Padding.Left,
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
