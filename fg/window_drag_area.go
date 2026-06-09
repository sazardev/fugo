package fg

import fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

// WindowDragAreaWidget turns its child into a handle that drags the OS window,
// the way a title bar does. It is meant for frameless windows where you draw
// your own chrome. Build one with WindowDragArea.
type WindowDragAreaWidget struct {
	child Widget
	baseWidget
}

// WindowDragArea wraps child so that pressing and dragging within it moves the
// application window.
func WindowDragArea(child Widget) *WindowDragAreaWidget {
	return &WindowDragAreaWidget{child: child}
}

func (w *WindowDragAreaWidget) isWidget() {}

func (w *WindowDragAreaWidget) widgetChildren() []Widget {
	if w.child != nil {
		return []Widget{w.child}
	}

	return nil
}

func (w *WindowDragAreaWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	w.id = *counter

	var (
		childIDs []uint32
		allNodes []*fugov1.WidgetNode
	)

	for _, child := range w.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	return append([]*fugov1.WidgetNode{{
		Id:       w.id,
		Key:      w.key,
		Type:     fugov1.WidgetType_WINDOWDRAGAREA,
		Children: childIDs,
	}}, allNodes...)
}
