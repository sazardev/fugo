package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AnimatedContainerWidget is a Container that animates changes to its visual properties. Build one with AnimatedContainer.
type AnimatedContainerWidget struct {
	child        Widget
	bgColor      style.Color
	Padding      style.EdgeInsets
	borderRadius float64
	durationMs   int32
	curve        string
	baseWidget
}

// AnimatedContainer creates an animated container wrapping child, with a default 200ms ease transition.
func AnimatedContainer(child Widget) *AnimatedContainerWidget {
	return &AnimatedContainerWidget{
		child:      child,
		curve:      "ease",
		durationMs: 200,
	}
}

// BgColor sets the target background color to animate toward and returns the widget for chaining.
func (c *AnimatedContainerWidget) BgColor(v style.Color) *AnimatedContainerWidget {
	c.bgColor = v

	return c
}

// Pad sets the target inner padding to animate toward and returns the widget for chaining.
func (c *AnimatedContainerWidget) Pad(v style.EdgeInsets) *AnimatedContainerWidget {
	c.Padding = v

	return c
}

// BorderRadius sets the target corner radius to animate toward and returns the widget for chaining.
func (c *AnimatedContainerWidget) BorderRadius(v float64) *AnimatedContainerWidget {
	c.borderRadius = v

	return c
}

// DurationMs sets the animation duration in milliseconds and returns the widget for chaining.
func (c *AnimatedContainerWidget) DurationMs(v int32) *AnimatedContainerWidget {
	c.durationMs = v

	return c
}

// Curve sets the animation easing curve by name and returns the widget for chaining.
func (c *AnimatedContainerWidget) Curve(v string) *AnimatedContainerWidget {
	c.curve = v

	return c
}

func (c *AnimatedContainerWidget) isWidget() {}

func (c *AnimatedContainerWidget) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *AnimatedContainerWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		BgColor:      c.bgColor.String(),
		Padding:      c.Padding.Top,
		BorderRadius: c.borderRadius,
		DurationMs:   c.durationMs,
		Curve:        c.curve,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_ANIMATEDCONTAINER,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
