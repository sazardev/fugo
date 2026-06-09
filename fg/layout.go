package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ColumnWidget arranges its children vertically with configurable axis
// alignment. Build one with Column. By default it shrinks to its children
// (MainAxisSize min) and centers them on both axes, which — combined with the
// renderer auto-centering a non-filling root — makes simple content land in
// the middle of the window. Call Expand to fill the vertical axis instead.
type ColumnWidget struct {
	items        []Widget
	mainAxisSize fugov1.MainAxisSize
	mainAlign    fugov1.MainAxisAlignment
	crossAlign   fugov1.CrossAxisAlignment
	baseWidget
}

// Column creates a vertical layout containing the given items.
func Column(items ...Widget) *ColumnWidget {
	return &ColumnWidget{
		items:        items,
		mainAxisSize: fugov1.MainAxisSize_MAIN_MIN,
		mainAlign:    fugov1.MainAxisAlignment_MAIN_CENTER,
		crossAlign:   fugov1.CrossAxisAlignment_CROSS_CENTER,
	}
}

// MainAlign sets how children are distributed along the vertical axis and returns the widget for chaining.
func (c *ColumnWidget) MainAlign(v fugov1.MainAxisAlignment) *ColumnWidget {
	c.mainAlign = v

	return c
}

// CrossAlign sets how children are aligned along the horizontal axis and returns the widget for chaining.
func (c *ColumnWidget) CrossAlign(v fugov1.CrossAxisAlignment) *ColumnWidget {
	c.crossAlign = v

	return c
}

// MainAxisSize sets whether the column shrinks to fit its children or fills the vertical axis, and returns the widget for chaining.
func (c *ColumnWidget) MainAxisSize(v fugov1.MainAxisSize) *ColumnWidget {
	c.mainAxisSize = v

	return c
}

// Expand makes the column fill the vertical axis (MainAxisSize max) and returns the widget for chaining.
func (c *ColumnWidget) Expand() *ColumnWidget {
	c.mainAxisSize = fugov1.MainAxisSize_MAIN_MAX

	return c
}

func (c *ColumnWidget) isWidget() {}

func (c *ColumnWidget) widgetChildren() []Widget { return c.items }

func (c *ColumnWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	childIDs, allNodes := walkChildren(c.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.ColumnProps{
		MainAxisSize:   c.mainAxisSize,
		MainAlignment:  c.mainAlign,
		CrossAlignment: c.crossAlign,
	})

	self := &fugov1.WidgetNode{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_COLUMN,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}

// CenterWidget centers its single child within the available space. Build one with Center.
type CenterWidget struct {
	child Widget
	baseWidget
}

// Center creates a widget that centers child.
func Center(child Widget) *CenterWidget {
	return &CenterWidget{child: child}
}

func (c *CenterWidget) isWidget() {}

func (c *CenterWidget) widgetChildren() []Widget {
	if c.child != nil {
		return []Widget{c.child}
	}

	return nil
}

func (c *CenterWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	c.id = *counter

	childIDs, allNodes := walkChildren(c.widgetChildren(), counter)

	self := &fugov1.WidgetNode{
		Id:       c.id,
		Key:      c.key,
		Type:     fugov1.WidgetType_CENTER,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
