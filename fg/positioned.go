package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type PositionedWidget struct {
	child  Widget
	left   float64
	top    float64
	right  float64
	bottom float64
	width  float64
	height float64
	baseWidget
}

func Positioned(child Widget) *PositionedWidget {
	return &PositionedWidget{child: child}
}

func (p *PositionedWidget) Left(v float64) *PositionedWidget {
	p.left = v

	return p
}

func (p *PositionedWidget) Top(v float64) *PositionedWidget {
	p.top = v

	return p
}

func (p *PositionedWidget) Right(v float64) *PositionedWidget {
	p.right = v

	return p
}

func (p *PositionedWidget) Bottom(v float64) *PositionedWidget {
	p.bottom = v

	return p
}

func (p *PositionedWidget) Width(v float64) *PositionedWidget {
	p.width = v

	return p
}

func (p *PositionedWidget) Height(v float64) *PositionedWidget {
	p.height = v

	return p
}

func (p *PositionedWidget) isWidget()                {}
func (p *PositionedWidget) widgetChildren() []Widget { return []Widget{p.child} }

func (p *PositionedWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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

	props, _ := proto.Marshal(&fugov1.PositionedProps{
		Left:   p.left,
		Top:    p.top,
		Right:  p.right,
		Bottom: p.bottom,
		Width:  p.width,
		Height: p.height,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       p.id,
		Key:      p.key,
		Type:     fugov1.WidgetType_POSITIONED,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
