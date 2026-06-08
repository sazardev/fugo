package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type ListViewWidget struct {
	items      []Widget
	itemExtent float64
	baseWidget
}

func ListView(items ...Widget) *ListViewWidget {
	return &ListViewWidget{items: items}
}

func (l *ListViewWidget) ItemExtent(v float64) *ListViewWidget {
	l.itemExtent = v

	return l
}

func (l *ListViewWidget) isWidget()                {}
func (l *ListViewWidget) widgetChildren() []Widget { return l.items }

func (l *ListViewWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	l.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range l.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.ListViewProps{
		ItemExtent: l.itemExtent,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       l.id,
		Key:      l.key,
		Type:     fugov1.WidgetType_LISTVIEW,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
