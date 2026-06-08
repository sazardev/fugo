package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type AnimatedContainer struct {
	child        Widget
	BgColor      style.Color
	Padding      style.EdgeInsets
	BorderRadius float64
	DurationMs   int32
	Curve        string
	baseWidget
}

func AnimatedContainer(child Widget) *AnimatedContainer {
	return &AnimatedContainer{
		child:      child,
		Curve:      "ease",
		DurationMs: 200,
	}
}

func (c *AnimatedContainer) BgColor(v style.Color) *AnimatedContainer {
	c.BgColor = v

	return c
}

func (c *AnimatedContainer) Pad(v style.EdgeInsets) *AnimatedContainer {
	c.Padding = v

	return c
}

func (c *AnimatedContainer) BorderRadius(v float64) *AnimatedContainer {
	c.BorderRadius = v

	return c
}

func (c *AnimatedContainer) DurationMs(v int32) *AnimatedContainer {
	c.DurationMs = v

	return c
}

func (c *AnimatedContainer) Curve(v string) *AnimatedContainer {
	c.Curve = v

	return c
}

func (c *AnimatedContainer) isWidget() {}

func (c *AnimatedContainer) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *AnimatedContainer) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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

	props, _ := proto.Marshal(&fugov1.AnimatedContainerProps{
		BgColor:      c.BgColor.String(),
		Padding:      c.Padding.Top,
		BorderRadius: c.BorderRadius,
		DurationMs:   c.DurationMs,
		Curve:        c.Curve,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_ANIMATEDCONTAINER,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
