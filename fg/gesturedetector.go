package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// GestureDetectorWidget makes any child tappable. Build one with GestureDetector.
type GestureDetectorWidget struct {
	child   Widget
	handler func(Event)
	baseWidget
}

// GestureDetector wraps child so taps invoke the handler registered with OnTap.
func GestureDetector(child Widget) *GestureDetectorWidget {
	return &GestureDetectorWidget{child: child}
}

// OnTap registers the tap handler and returns the widget for chaining.
func (g *GestureDetectorWidget) OnTap(handler func(Event)) *GestureDetectorWidget {
	g.handler = handler

	return g
}

func (g *GestureDetectorWidget) isWidget() {}

func (g *GestureDetectorWidget) widgetChildren() []Widget {
	if g.child != nil {
		return []Widget{g.child}
	}

	return nil
}

// HasHandler reports whether an OnTap handler has been registered.
func (g *GestureDetectorWidget) HasHandler() bool { return g.handler != nil }

// Handle dispatches event to the registered OnTap handler, if any.
func (g *GestureDetectorWidget) Handle(event Event) {
	if g.handler != nil {
		g.handler(event)
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

	return append([]*fugov1.WidgetNode{{
		Id:       g.id,
		Key:      g.key,
		Type:     fugov1.WidgetType_GESTUREDETECTOR,
		Children: childIDs,
	}}, allNodes...)
}
