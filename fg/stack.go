package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// StackWidget overlays its children on top of one another. Build one with Stack.
type StackWidget struct {
	items []Widget
	baseWidget
}

// Stack creates a layout that stacks the given items, later items drawn on top.
func Stack(items ...Widget) *StackWidget {
	return &StackWidget{items: items}
}

func (s *StackWidget) isWidget()                {}
func (s *StackWidget) widgetChildren() []Widget { return s.items }

func (s *StackWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range s.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	return append([]*fugov1.WidgetNode{{
		Id:       s.id,
		Key:      s.key,
		Type:     fugov1.WidgetType_STACK,
		Children: childIDs,
	}}, allNodes...)
}
