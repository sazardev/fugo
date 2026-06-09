package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AnimatedPositionedWidget places its child at explicit offsets within a Stack
// and animates changes to those offsets. Build one with AnimatedPositioned; it
// must be a direct child of a Stack, like Positioned.
type AnimatedPositionedWidget struct {
	child      Widget
	curve      string
	left       float64
	top        float64
	right      float64
	bottom     float64
	width      float64
	height     float64
	durationMs int32
	baseWidget
}

// AnimatedPositioned creates an animated positioned wrapper around child, with
// a default 200ms ease transition. Mutate its offsets and call Context.Update
// to animate the child gliding to the new position.
func AnimatedPositioned(child Widget) *AnimatedPositionedWidget {
	return &AnimatedPositionedWidget{
		child:      child,
		curve:      "ease",
		durationMs: 200,
	}
}

// Left sets the target offset from the stack's left edge and returns the widget for chaining.
func (p *AnimatedPositionedWidget) Left(v float64) *AnimatedPositionedWidget {
	p.left = v

	return p
}

// Top sets the target offset from the stack's top edge and returns the widget for chaining.
func (p *AnimatedPositionedWidget) Top(v float64) *AnimatedPositionedWidget {
	p.top = v

	return p
}

// Right sets the target offset from the stack's right edge and returns the widget for chaining.
func (p *AnimatedPositionedWidget) Right(v float64) *AnimatedPositionedWidget {
	p.right = v

	return p
}

// Bottom sets the target offset from the stack's bottom edge and returns the widget for chaining.
func (p *AnimatedPositionedWidget) Bottom(v float64) *AnimatedPositionedWidget {
	p.bottom = v

	return p
}

// Width sets the target width in logical pixels and returns the widget for chaining.
func (p *AnimatedPositionedWidget) Width(v float64) *AnimatedPositionedWidget {
	p.width = v

	return p
}

// Height sets the target height in logical pixels and returns the widget for chaining.
func (p *AnimatedPositionedWidget) Height(v float64) *AnimatedPositionedWidget {
	p.height = v

	return p
}

// DurationMs sets the animation duration in milliseconds and returns the widget for chaining.
func (p *AnimatedPositionedWidget) DurationMs(v int32) *AnimatedPositionedWidget {
	p.durationMs = v

	return p
}

// Curve sets the animation easing curve by name and returns the widget for chaining.
func (p *AnimatedPositionedWidget) Curve(v string) *AnimatedPositionedWidget {
	p.curve = v

	return p
}

func (p *AnimatedPositionedWidget) isWidget()                {}
func (p *AnimatedPositionedWidget) widgetChildren() []Widget { return []Widget{p.child} }

func (p *AnimatedPositionedWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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

	props, _ := proto.Marshal(&fugov1.AnimatedPositionedProps{
		Left:       p.left,
		Top:        p.top,
		Right:      p.right,
		Bottom:     p.bottom,
		Width:      p.width,
		Height:     p.height,
		DurationMs: p.durationMs,
		Curve:      p.curve,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       p.id,
		Key:      p.key,
		Type:     fugov1.WidgetType_ANIMATEDPOSITIONED,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
