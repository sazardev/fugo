package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AlignWidget positions its child within itself using fractional alignment.
// Build one with Align.
type AlignWidget struct {
	child Widget
	x     float64
	y     float64
	baseWidget
}

// Align wraps child, aligning it at (x, y) where each axis runs from -1
// (start/top) to 1 (end/bottom); (0, 0) is centered.
func Align(child Widget, x, y float64) *AlignWidget {
	return &AlignWidget{child: child, x: x, y: y}
}

func (a *AlignWidget) isWidget() {}

func (a *AlignWidget) widgetChildren() []Widget {
	if a.child != nil {
		return []Widget{a.child}
	}

	return nil
}

func (a *AlignWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	a.id = *counter

	var (
		childIDs []uint32
		allNodes []*fugov1.WidgetNode
	)

	for _, child := range a.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.AlignProps{X: a.x, Y: a.y})

	return append([]*fugov1.WidgetNode{{
		Id:       a.id,
		Key:      a.key,
		Type:     fugov1.WidgetType_ALIGN,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
