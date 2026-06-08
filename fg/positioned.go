package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Positioned struct {
	child  Widget
	Left   float64
	Top    float64
	Right  float64
	Bottom float64
	Width  float64
	Height float64
	baseWidget
}

func Positioned(child Widget) *Positioned {
	return &Positioned{child: child}
}

func (p *Positioned) Left(v float64) *Positioned {
	p.Left = v

	return p
}

func (p *Positioned) Top(v float64) *Positioned {
	p.Top = v

	return p
}

func (p *Positioned) Right(v float64) *Positioned {
	p.Right = v

	return p
}

func (p *Positioned) Bottom(v float64) *Positioned {
	p.Bottom = v

	return p
}

func (p *Positioned) Width(v float64) *Positioned {
	p.Width = v

	return p
}

func (p *Positioned) Height(v float64) *Positioned {
	p.Height = v

	return p
}

func (p *Positioned) isWidget()                {}
func (p *Positioned) widgetChildren() []Widget { return []Widget{p.child} }

func (p *Positioned) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		Left:   p.Left,
		Top:    p.Top,
		Right:  p.Right,
		Bottom: p.Bottom,
		Width:  p.Width,
		Height: p.Height,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       p.id,
		Key:      p.key,
		Type:     fugov1.WidgetType_POSITIONED,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
