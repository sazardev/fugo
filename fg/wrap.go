package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// WrapWidget lays out its children in a row that wraps onto new runs when space runs out. Build one with Wrap.
type WrapWidget struct {
	children   []Widget
	spacing    float64
	runSpacing float64
	baseWidget
}

// Wrap creates a wrapping layout containing the given children.
func Wrap(children ...Widget) *WrapWidget {
	return &WrapWidget{children: children}
}

// Spacing sets the gap between children within a run and returns the widget for chaining.
func (w *WrapWidget) Spacing(v float64) *WrapWidget {
	w.spacing = v

	return w
}

// RunSpacing sets the gap between successive runs and returns the widget for chaining.
func (w *WrapWidget) RunSpacing(v float64) *WrapWidget {
	w.runSpacing = v

	return w
}

func (w *WrapWidget) isWidget()                {}
func (w *WrapWidget) widgetChildren() []Widget { return w.children }

func (w *WrapWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	w.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range w.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.WrapProps{
		Spacing:    w.spacing,
		RunSpacing: w.runSpacing,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       w.id,
		Key:      w.key,
		Type:     fugov1.WidgetType_WRAP,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
