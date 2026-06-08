package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// SizedBoxWidget is a box with a fixed width and height, optionally wrapping a child. Build one with SizedBox.
type SizedBoxWidget struct {
	child  Widget
	width  float64
	height float64
	baseWidget
}

// SizedBox creates a box of the given width and height in logical pixels.
func SizedBox(width, height float64) *SizedBoxWidget {
	return &SizedBoxWidget{width: width, height: height}
}

// Child sets the widget placed inside the box and returns the widget for chaining.
func (s *SizedBoxWidget) Child(child Widget) *SizedBoxWidget {
	s.child = child

	return s
}

// Width sets the box width in logical pixels and returns the widget for chaining.
func (s *SizedBoxWidget) Width(v float64) *SizedBoxWidget {
	s.width = v

	return s
}

// Height sets the box height in logical pixels and returns the widget for chaining.
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
