package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ConstrainedBoxWidget imposes additional min/max size constraints on its
// child. Build one with ConstrainedBox.
type ConstrainedBoxWidget struct {
	child                                    Widget
	minWidth, maxWidth, minHeight, maxHeight float64
	baseWidget
}

// ConstrainedBox wraps child with size constraints; use MinWidth/MaxWidth/
// MinHeight/MaxHeight to set them (0 means unconstrained on that field).
func ConstrainedBox(child Widget) *ConstrainedBoxWidget {
	return &ConstrainedBoxWidget{child: child}
}

// MinWidth sets the minimum width constraint and returns the widget for chaining.
func (c *ConstrainedBoxWidget) MinWidth(v float64) *ConstrainedBoxWidget {
	c.minWidth = v

	return c
}

// MaxWidth sets the maximum width constraint and returns the widget for chaining.
func (c *ConstrainedBoxWidget) MaxWidth(v float64) *ConstrainedBoxWidget {
	c.maxWidth = v

	return c
}

// MinHeight sets the minimum height constraint and returns the widget for chaining.
func (c *ConstrainedBoxWidget) MinHeight(v float64) *ConstrainedBoxWidget {
	c.minHeight = v

	return c
}

// MaxHeight sets the maximum height constraint and returns the widget for chaining.
func (c *ConstrainedBoxWidget) MaxHeight(v float64) *ConstrainedBoxWidget {
	c.maxHeight = v

	return c
}

func (c *ConstrainedBoxWidget) isWidget()                {}
func (c *ConstrainedBoxWidget) widgetChildren() []Widget { return oneChild(c.child) }

func (c *ConstrainedBoxWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter
	ids, nodes := walkChildren(c.widgetChildren(), counter)
	props, err := proto.Marshal(&fugov1.ConstrainedBoxProps{
		MinWidth:  c.minWidth,
		MaxWidth:  c.maxWidth,
		MinHeight: c.minHeight,
		MaxHeight: c.maxHeight,
	})
	if err != nil {
		flog.Errorf("marshal ConstrainedBoxProps: %v", err)
	}

	return selfNode(c.id, c.key, fugov1.WidgetType_CONSTRAINEDBOX, props, ids, nodes)
}

// FractionallySizedBoxWidget sizes its child to a fraction of the available
// space along one or both axes. Build one with FractionallySizedBox.
type FractionallySizedBoxWidget struct {
	child                     Widget
	widthFactor, heightFactor float64
	baseWidget
}

// FractionallySizedBox wraps child so it can be sized as a fraction of the
// available space; use WidthFactor/HeightFactor to set them (0 means that
// axis is unconstrained).
func FractionallySizedBox(child Widget) *FractionallySizedBoxWidget {
	return &FractionallySizedBoxWidget{child: child}
}

// WidthFactor sets the fraction of the available width and returns the widget for chaining.
func (f *FractionallySizedBoxWidget) WidthFactor(v float64) *FractionallySizedBoxWidget {
	f.widthFactor = v

	return f
}

// HeightFactor sets the fraction of the available height and returns the widget for chaining.
func (f *FractionallySizedBoxWidget) HeightFactor(v float64) *FractionallySizedBoxWidget {
	f.heightFactor = v

	return f
}

func (f *FractionallySizedBoxWidget) isWidget()                {}
func (f *FractionallySizedBoxWidget) widgetChildren() []Widget { return oneChild(f.child) }

func (f *FractionallySizedBoxWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	f.id = *counter
	ids, nodes := walkChildren(f.widgetChildren(), counter)
	props, err := proto.Marshal(&fugov1.FractionallySizedBoxProps{
		WidthFactor:  f.widthFactor,
		HeightFactor: f.heightFactor,
	})
	if err != nil {
		flog.Errorf("marshal FractionallySizedBoxProps: %v", err)
	}

	return selfNode(f.id, f.key, fugov1.WidgetType_FRACTIONALLYSIZEDBOX, props, ids, nodes)
}
