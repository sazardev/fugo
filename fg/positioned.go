package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// PositionedWidget places its child at explicit offsets within a Stack. Build one with Positioned.
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

// Positioned creates a positioned wrapper around child for use inside a Stack.
func Positioned(child Widget) *PositionedWidget {
	return &PositionedWidget{child: child}
}

// Left sets the offset from the stack's left edge and returns the widget for chaining.
func (p *PositionedWidget) Left(v float64) *PositionedWidget {
	p.left = v

	return p
}

// Top sets the offset from the stack's top edge and returns the widget for chaining.
func (p *PositionedWidget) Top(v float64) *PositionedWidget {
	p.top = v

	return p
}

// Right sets the offset from the stack's right edge and returns the widget for chaining.
func (p *PositionedWidget) Right(v float64) *PositionedWidget {
	p.right = v

	return p
}

// Bottom sets the offset from the stack's bottom edge and returns the widget for chaining.
func (p *PositionedWidget) Bottom(v float64) *PositionedWidget {
	p.bottom = v

	return p
}

// Width sets the child's width in logical pixels and returns the widget for chaining.
func (p *PositionedWidget) Width(v float64) *PositionedWidget {
	p.width = v

	return p
}

// Height sets the child's height in logical pixels and returns the widget for chaining.
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
