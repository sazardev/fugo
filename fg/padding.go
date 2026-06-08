package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type PaddingWidget struct {
	child  Widget
	Insets EdgeInsets
	baseWidget
}

func Padding(child Widget, insets EdgeInsets) *PaddingWidget {
	return &PaddingWidget{child: child, Insets: insets}
}

func PaddingAll(child Widget, value float64) *PaddingWidget {
	return Padding(child, EdgeAll(value))
}

func (p *PaddingWidget) isWidget() {}
func (p *PaddingWidget) widgetChildren() []Widget {
	if p.child != nil {
		return []Widget{p.child}
	}

	return nil
}

func (p *PaddingWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	p.id = *counter

	var (
		childIDs []uint32
		allNodes []*fugov1.WidgetNode
	)

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
