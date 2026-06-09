package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AspectRatioWidget sizes its child to a width:height ratio. Build with AspectRatio.
type AspectRatioWidget struct {
	child Widget
	ratio float64
	baseWidget
}

// AspectRatio constrains child to the given width/height ratio (e.g. 16.0/9.0).
func AspectRatio(ratio float64, child Widget) *AspectRatioWidget {
	return &AspectRatioWidget{ratio: ratio, child: child}
}

func (a *AspectRatioWidget) isWidget()                {}
func (a *AspectRatioWidget) widgetChildren() []Widget { return oneChild(a.child) }

func (a *AspectRatioWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	a.id = *counter
	ids, nodes := walkChildren(a.widgetChildren(), counter)
	props, _ := proto.Marshal(&fugov1.AspectRatioProps{Ratio: a.ratio})

	return selfNode(a.id, a.key, fugov1.WidgetType_ASPECTRATIO, props, ids, nodes)
}

// ClipRRectWidget clips its child with rounded corners. Build with ClipRRect.
type ClipRRectWidget struct {
	child  Widget
	radius float64
	baseWidget
}

// ClipRRect clips child to a rounded rectangle of the given corner radius.
func ClipRRect(radius float64, child Widget) *ClipRRectWidget {
	return &ClipRRectWidget{radius: radius, child: child}
}

func (c *ClipRRectWidget) isWidget()                {}
func (c *ClipRRectWidget) widgetChildren() []Widget { return oneChild(c.child) }

func (c *ClipRRectWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter
	ids, nodes := walkChildren(c.widgetChildren(), counter)
	props, _ := proto.Marshal(&fugov1.ClipRRectProps{Radius: c.radius})

	return selfNode(c.id, c.key, fugov1.WidgetType_CLIPRRECT, props, ids, nodes)
}

// FittedBoxWidget scales its child to fit the available space. Build with FittedBox.
type FittedBoxWidget struct {
	child Widget
	baseWidget
}

// FittedBox scales child down to fit (BoxFit.contain).
func FittedBox(child Widget) *FittedBoxWidget {
	return &FittedBoxWidget{child: child}
}

func (f *FittedBoxWidget) isWidget()                {}
func (f *FittedBoxWidget) widgetChildren() []Widget { return oneChild(f.child) }

func (f *FittedBoxWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	f.id = *counter
	ids, nodes := walkChildren(f.widgetChildren(), counter)
	props, _ := proto.Marshal(&fugov1.FittedBoxProps{})

	return selfNode(f.id, f.key, fugov1.WidgetType_FITTEDBOX, props, ids, nodes)
}

// FlexibleWidget lets its child take a flexible share of a Row/Column without
// forcing it to fill (the loose-fit counterpart of Expanded). Build with Flexible.
type FlexibleWidget struct {
	child Widget
	flex  int
	baseWidget
}

// Flexible wraps child so it may take up to flex 1 of the free space.
func Flexible(child Widget) *FlexibleWidget {
	return &FlexibleWidget{child: child, flex: 1}
}

// Flex sets this child's flex factor and returns the widget for chaining.
func (f *FlexibleWidget) Flex(n int) *FlexibleWidget {
	f.flex = n

	return f
}

func (f *FlexibleWidget) isWidget()                {}
func (f *FlexibleWidget) widgetChildren() []Widget { return oneChild(f.child) }

func (f *FlexibleWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	f.id = *counter
	ids, nodes := walkChildren(f.widgetChildren(), counter)
	props, _ := proto.Marshal(&fugov1.FlexibleProps{
		Flex: int32(f.flex), //nolint:gosec // a small flex factor
	})

	return selfNode(f.id, f.key, fugov1.WidgetType_FLEXIBLE, props, ids, nodes)
}
