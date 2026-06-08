package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type SizedBoxWidget struct {
	child  Widget
	width  float64
	height float64
	baseWidget
}

func SizedBox(width, height float64) *SizedBoxWidget {
	return &SizedBoxWidget{width: width, height: height}
}

func (s *SizedBoxWidget) Child(child Widget) *SizedBoxWidget {
	s.child = child

	return s
}

func (s *SizedBoxWidget) Width(v float64) *SizedBoxWidget {
	s.width = v

	return s
}

func (s *SizedBoxWidget) Height(v float64) *SizedBoxWidget {
	s.height = v

	return s
}

func (s *SizedBoxWidget) isWidget() {}
func (s *SizedBoxWidget) widgetChildren() []Widget {
	if s.child != nil {
		return []Widget{s.child}
	}

	return nil
}

func (s *SizedBoxWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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

	props, _ := proto.Marshal(&fugov1.SizedBoxProps{
		Width:  s.width,
		Height: s.height,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       s.id,
		Key:      s.key,
		Type:     fugov1.WidgetType_SIZEDBOX,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
