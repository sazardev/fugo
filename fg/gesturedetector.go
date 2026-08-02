package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// GestureDetectorWidget makes any child tappable, draggable, and/or
// scalable. Build one with GestureDetector.
type GestureDetectorWidget struct {
	child         Widget
	onTap         func(Event)
	onDoubleTap   func(Event)
	onLongPress   func(Event)
	onPanStart    func(Event)
	onPanUpdate   func(Event) // Event.Data = "dx,dy" delta since last update, e.g. "3.50,-1.20"
	onPanEnd      func(Event)
	onScaleStart  func(Event)
	onScaleUpdate func(Event) // Event.Data = "scale,dx,dy", e.g. "1.25,3.50,-1.20"
	onScaleEnd    func(Event)
	baseWidget
}

// GestureDetector wraps child so gestures invoke the handlers registered via
// OnTap, OnDoubleTap, OnLongPress, OnPan*, and OnScale*.
func GestureDetector(child Widget) *GestureDetectorWidget {
	return &GestureDetectorWidget{child: child}
}

// OnTap registers the tap handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnTap(handler func(Event)) *GestureDetectorWidget {
	g.onTap = handler

	return g
}

// OnDoubleTap registers the double-tap handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnDoubleTap(handler func(Event)) *GestureDetectorWidget {
	g.onDoubleTap = handler

	return g
}

// OnLongPress registers the long-press handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnLongPress(handler func(Event)) *GestureDetectorWidget {
	g.onLongPress = handler

	return g
}

// OnPanStart registers the pan-start handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnPanStart(handler func(Event)) *GestureDetectorWidget {
	g.onPanStart = handler

	return g
}

// OnPanUpdate registers the pan-update handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnPanUpdate(handler func(Event)) *GestureDetectorWidget {
	g.onPanUpdate = handler

	return g
}

// OnPanEnd registers the pan-end handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnPanEnd(handler func(Event)) *GestureDetectorWidget {
	g.onPanEnd = handler

	return g
}

// OnScaleStart registers the scale-start handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnScaleStart(handler func(Event)) *GestureDetectorWidget {
	g.onScaleStart = handler

	return g
}

// OnScaleUpdate registers the scale-update handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnScaleUpdate(handler func(Event)) *GestureDetectorWidget {
	g.onScaleUpdate = handler

	return g
}

// OnScaleEnd registers the scale-end handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnScaleEnd(handler func(Event)) *GestureDetectorWidget {
	g.onScaleEnd = handler

	return g
}

func (g *GestureDetectorWidget) isWidget() {}

func (g *GestureDetectorWidget) widgetChildren() []Widget {
	if g.child != nil {
		return []Widget{g.child}
	}

	return nil
}

// HasHandler reports whether any gesture handler has been registered.
func (g *GestureDetectorWidget) HasHandler() bool {
	return g.onTap != nil ||
		g.onDoubleTap != nil ||
		g.onLongPress != nil ||
		g.onPanStart != nil ||
		g.onPanUpdate != nil ||
		g.onPanEnd != nil ||
		g.onScaleStart != nil ||
		g.onScaleUpdate != nil ||
		g.onScaleEnd != nil
}

// Handle dispatches event to the registered handler matching event.EventType, if any.
func (g *GestureDetectorWidget) Handle(event Event) {
	var h func(Event)

	switch event.EventType {
	case "onTap":
		h = g.onTap
	case "onDoubleTap":
		h = g.onDoubleTap
	case "onLongPress":
		h = g.onLongPress
	case "onPanStart":
		h = g.onPanStart
	case "onPanUpdate":
		h = g.onPanUpdate
	case "onPanEnd":
		h = g.onPanEnd
	case "onScaleStart":
		h = g.onScaleStart
	case "onScaleUpdate":
		h = g.onScaleUpdate
	case "onScaleEnd":
		h = g.onScaleEnd
	}

	if h != nil {
		h(event)
	}
}

func (g *GestureDetectorWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	g.id = *counter

	var (
		childIDs []uint32
		allNodes []*fugov1.WidgetNode
	)

	for _, child := range g.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, err := proto.Marshal(&fugov1.GestureDetectorProps{
		HasDoubleTap: g.onDoubleTap != nil,
		HasLongPress: g.onLongPress != nil,
		HasPan:       g.onPanStart != nil || g.onPanUpdate != nil || g.onPanEnd != nil,
		HasScale:     g.onScaleStart != nil || g.onScaleUpdate != nil || g.onScaleEnd != nil,
	})
	if err != nil {
		flog.Errorf("marshal GestureDetectorProps: %v", err)
	}

	return append([]*fugov1.WidgetNode{{
		Id:       g.id,
		Key:      g.key,
		Type:     fugov1.WidgetType_GESTUREDETECTOR,
		Children: childIDs,
		Props:    props,
	}}, allNodes...)
}
