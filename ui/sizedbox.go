package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type SizedBox struct {
	child  Widget
	Width  float64
	Height float64
	baseWidget
}

func NewSizedBox(width, height float64) *SizedBox {
	return &SizedBox{Width: width, Height: height}
}

func (s *SizedBox) WithChild(child Widget) *SizedBox {
	s.child = child

	return s
}

func (s *SizedBox) WithWidth(v float64) *SizedBox {
	s.Width = v

	return s
}

func (s *SizedBox) WithHeight(v float64) *SizedBox {
	s.Height = v

	return s
}

func (s *SizedBox) isWidget() {}
func (s *SizedBox) widgetChildren() []Widget {
	if s.child != nil {
		return []Widget{s.child}
	}

	return nil
}

func (s *SizedBox) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		Width:  s.Width,
		Height: s.Height,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       s.id,
		Key:      s.key,
		Type:     fugov1.WidgetType_SIZEDBOX,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
