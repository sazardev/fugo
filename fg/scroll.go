package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ScrollViewWidget makes a single child scrollable along one axis. Build one
// with ScrollView (vertical by default; call Horizontal to switch axis).
type ScrollViewWidget struct {
	child     Widget
	direction int32
	baseWidget
}

// ScrollView wraps child in a vertically scrollable view.
func ScrollView(child Widget) *ScrollViewWidget {
	return &ScrollViewWidget{child: child}
}

// Horizontal switches the scroll axis to horizontal and returns the widget for chaining.
func (s *ScrollViewWidget) Horizontal() *ScrollViewWidget {
	s.direction = 1

	return s
}

func (s *ScrollViewWidget) isWidget() {}

func (s *ScrollViewWidget) widgetChildren() []Widget {
	if s.child != nil {
		return []Widget{s.child}
	}

	return nil
}

func (s *ScrollViewWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	var (
		childIDs []uint32
		allNodes []*fugov1.WidgetNode
	)

	for _, child := range s.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.ScrollViewProps{ScrollDirection: s.direction})

	return append([]*fugov1.WidgetNode{{
		Id:       s.id,
		Key:      s.key,
		Type:     fugov1.WidgetType_SCROLLVIEW,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
