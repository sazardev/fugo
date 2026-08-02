package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// CardWidget is a Material 3 card surface wrapping a single child. Build one
// with Card. Elevation, padding and corner radius fall back to the M3 defaults
// unless set.
type CardWidget struct {
	child        Widget
	elevation    float64
	padding      float64
	borderRadius float64
	baseWidget
}

// Card creates a Material card wrapping child.
func Card(child Widget) *CardWidget {
	return &CardWidget{child: child}
}

// Elevation sets the card's shadow elevation in logical pixels and returns the widget for chaining.
func (c *CardWidget) Elevation(v float64) *CardWidget {
	c.elevation = v

	return c
}

// Pad sets uniform inner padding in logical pixels and returns the widget for chaining.
func (c *CardWidget) Pad(v float64) *CardWidget {
	c.padding = v

	return c
}

// BorderRadius sets the corner radius in logical pixels and returns the widget for chaining.
func (c *CardWidget) BorderRadius(v float64) *CardWidget {
	c.borderRadius = v

	return c
}

func (c *CardWidget) isWidget() {}

func (c *CardWidget) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *CardWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	childIDs, allNodes := walkChildren(c.widgetChildren(), counter)

	// A zero elevation/radius means the caller never called Elevation/
	// BorderRadius explicitly (0 is never a value someone actually wants for
	// either), so fall back to the active theme's per-component defaults.
	elevation := c.elevation
	if elevation == 0 {
		elevation = CurrentTheme().Components.CardElevation
	}

	borderRadius := c.borderRadius
	if borderRadius == 0 {
		borderRadius = CurrentTheme().Components.CardRadius
	}

	props, _ := proto.Marshal(&fugov1.CardProps{
		Elevation:    elevation,
		Padding:      c.padding,
		BorderRadius: borderRadius,
	})

	self := &fugov1.WidgetNode{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_CARD,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
