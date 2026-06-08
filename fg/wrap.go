package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Wrap struct {
	children   []Widget
	Spacing    float64
	RunSpacing float64
	baseWidget
}

func Wrap(children ...Widget) *Wrap {
	return &Wrap{children: children}
}

func (w *Wrap) Spacing(v float64) *Wrap {
	w.Spacing = v

	return w
}

func (w *Wrap) RunSpacing(v float64) *Wrap {
	w.RunSpacing = v

	return w
}

func (w *Wrap) isWidget()                {}
func (w *Wrap) widgetChildren() []Widget { return w.children }

func (w *Wrap) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		Spacing:    w.Spacing,
		RunSpacing: w.RunSpacing,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       w.id,
		Key:      w.key,
		Type:     fugov1.WidgetType_WRAP,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
