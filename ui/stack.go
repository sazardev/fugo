package ui

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

type Stack struct {
	items []Widget
	baseWidget
}

func NewStack(items ...Widget) *Stack {
	return &Stack{items: items}
}

func (s *Stack) isWidget()                {}
func (s *Stack) widgetChildren() []Widget { return s.items }

func (s *Stack) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
