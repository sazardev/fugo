package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ContainerWidget wraps a single child with a background, padding, margin,
// rounded corners, border, and shadow. Build one with Container.
type ContainerWidget struct {
	child          Widget
	Padding        style.EdgeInsets
	margin         style.EdgeInsets
	borderRadius   float64
	bgColor        style.Color
	bgColorSet     bool
	borderColor    style.Color
	borderColorSet bool
	borderWidth    float64
	shadowColor    style.Color
	shadowColorSet bool
	shadowBlur     float64
	shadowOffsetX  float64
	shadowOffsetY  float64
	baseWidget
}

// Container creates a container wrapping child. It is transparent unless a
// background is set with BgColor (so it inherits the Material 3 surface).
func Container(child Widget) *ContainerWidget {
	return &ContainerWidget{child: child}
}

// BgColor sets the background color and returns the widget for chaining.
func (c *ContainerWidget) BgColor(v style.Color) *ContainerWidget {
	c.bgColor = v
	c.bgColorSet = true

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

// Margin sets the outer margin and returns the widget for chaining.
func (c *ContainerWidget) Margin(v style.EdgeInsets) *ContainerWidget {
	c.margin = v

	return c
}

// Border sets a solid border color and width (in logical pixels) and returns
// the widget for chaining. If width is 0 and color is non-empty, width
// defaults to 1.0 (a hairline border) so a caller doing Border(color, 0)
// still gets a visible border instead of an invisible one.
func (c *ContainerWidget) Border(color style.Color, width float64) *ContainerWidget {
	c.borderColor = color
	c.borderColorSet = true

	if width == 0 {
		width = 1.0
	}

	c.borderWidth = width

	return c
}

// Shadow sets a drop shadow (color, blur radius, and offset in logical
// pixels) and returns the widget for chaining.
func (c *ContainerWidget) Shadow(color style.Color, blur, offsetX, offsetY float64) *ContainerWidget {
	c.shadowColor = color
	c.shadowColorSet = true
	c.shadowBlur = blur
	c.shadowOffsetX = offsetX
	c.shadowOffsetY = offsetY

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

	childIDs, allNodes := walkChildren(c.widgetChildren(), counter)

	bgColor := ""
	if c.bgColorSet {
		bgColor = c.bgColor.String()
	}

	borderColor := ""
	if c.borderColorSet {
		borderColor = c.borderColor.String()
	}

	shadowColor := ""
	if c.shadowColorSet {
		shadowColor = c.shadowColor.String()
	}

	props, _ := proto.Marshal(&fugov1.ContainerProps{
		BgColor:       bgColor,
		BorderRadius:  c.borderRadius,
		PadTop:        c.Padding.Top,
		PadRight:      c.Padding.Right,
		PadBottom:     c.Padding.Bottom,
		PadLeft:       c.Padding.Left,
		MarginTop:     c.margin.Top,
		MarginRight:   c.margin.Right,
		MarginBottom:  c.margin.Bottom,
		MarginLeft:    c.margin.Left,
		BorderColor:   borderColor,
		BorderWidth:   c.borderWidth,
		ShadowColor:   shadowColor,
		ShadowBlur:    c.shadowBlur,
		ShadowOffsetX: c.shadowOffsetX,
		ShadowOffsetY: c.shadowOffsetY,
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
