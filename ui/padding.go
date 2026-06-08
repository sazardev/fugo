package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Padding struct {
	child  Widget
	Insets struct{ Top, Right, Bottom, Left float64 }
	baseWidget
}

func NewPadding(child Widget, top, right, bottom, left float64) *Padding {
	return &Padding{
		child: child,
		Insets: struct {
			Top    float64
			Right  float64
			Bottom float64
			Left   float64
		}{Top: top, Right: right, Bottom: bottom, Left: left},
	}
}

func PaddingAll(child Widget, value float64) *Padding {
	return NewPadding(child, value, value, value, value)
}

func (p *Padding) isWidget() {}
func (p *Padding) widgetChildren() []Widget {
	if p.child != nil {
		return []Widget{p.child}
	}

	return nil
}

func (p *Padding) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	p.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range p.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.PaddingProps{
		Top:    p.Insets.Top,
		Right:  p.Insets.Right,
		Bottom: p.Insets.Bottom,
		Left:   p.Insets.Left,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       p.id,
		Key:      p.key,
		Type:     fugov1.WidgetType_PADDING,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
