package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Expanded struct {
	child Widget
	Flex  int32
	baseWidget
}

func NewExpanded(child Widget) *Expanded {
	return &Expanded{child: child, Flex: 1}
}

func (e *Expanded) WithFlex(v int32) *Expanded {
	e.Flex = v

	return e
}

func (e *Expanded) isWidget() {}
func (e *Expanded) widgetChildren() []Widget {
	if e.child != nil {
		return []Widget{e.child}
	}

	return nil
}

func (e *Expanded) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	e.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range e.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.ExpandedProps{Flex: e.Flex})

	return append([]*fugov1.WidgetNode{{
		Id:       e.id,
		Key:      e.key,
		Type:     fugov1.WidgetType_EXPANDED,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
