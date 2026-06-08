package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type GridView struct {
	children         []Widget
	CrossAxisCount   int32
	ChildAspectRatio float64
	baseWidget
}

func NewGridView(children ...Widget) *GridView {
	return &GridView{
		children:       children,
		CrossAxisCount: 2,
	}
}

func (g *GridView) WithCrossAxisCount(v int32) *GridView {
	g.CrossAxisCount = v

	return g
}

func (g *GridView) WithChildAspectRatio(v float64) *GridView {
	g.ChildAspectRatio = v

	return g
}

func (g *GridView) isWidget()                {}
func (g *GridView) widgetChildren() []Widget { return g.children }

func (g *GridView) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		CrossAxisCount:   g.CrossAxisCount,
		ChildAspectRatio: g.ChildAspectRatio,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       g.id,
		Key:      g.key,
		Type:     fugov1.WidgetType_GRIDVIEW,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
