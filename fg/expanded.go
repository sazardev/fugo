package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type ExpandedWidget struct {
	child Widget
	flex  int32
	baseWidget
}

func Expanded(child Widget) *ExpandedWidget {
	return &ExpandedWidget{child: child, flex: 1}
}

func (e *ExpandedWidget) Flex(v int32) *ExpandedWidget {
	e.flex = v

	return e
}

func (e *ExpandedWidget) isWidget() {}
func (e *ExpandedWidget) widgetChildren() []Widget {
	if e.child != nil {
		return []Widget{e.child}
	}

	return nil
}

func (e *ExpandedWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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

	props, _ := proto.Marshal(&fugov1.ExpandedProps{Flex: e.flex})

	return append([]*fugov1.WidgetNode{{
		Id:       e.id,
		Key:      e.key,
		Type:     fugov1.WidgetType_EXPANDED,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
