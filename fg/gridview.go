package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// GridViewWidget arranges its children in a scrollable grid. Build one with GridView.
type GridViewWidget struct {
	children         []Widget
	crossAxisCount   int32
	childAspectRatio float64
	baseWidget
}

// GridView creates a grid of the given children with two columns by default.
func GridView(children ...Widget) *GridViewWidget {
	return &GridViewWidget{
		children:       children,
		crossAxisCount: 2,
	}
}

// CrossAxisCount sets the number of columns and returns the widget for chaining.
func (g *GridViewWidget) CrossAxisCount(v int32) *GridViewWidget {
	g.crossAxisCount = v

	return g
}

// ChildAspectRatio sets the width-to-height ratio of each grid cell and returns the widget for chaining.
func (g *GridViewWidget) ChildAspectRatio(v float64) *GridViewWidget {
	g.childAspectRatio = v

	return g
}

func (g *GridViewWidget) isWidget()                {}
func (g *GridViewWidget) widgetChildren() []Widget { return g.children }

func (g *GridViewWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	g.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range g.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.GridViewProps{
		CrossAxisCount:   g.crossAxisCount,
		ChildAspectRatio: g.childAspectRatio,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       g.id,
		Key:      g.key,
		Type:     fugov1.WidgetType_GRIDVIEW,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
