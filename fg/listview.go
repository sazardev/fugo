package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type ListView struct {
	items      []Widget
	ItemExtent float64
	baseWidget
}

func NewListView(items ...Widget) *ListView {
	return &ListView{items: items}
}

func (l *ListView) WithItemExtent(v float64) *ListView {
	l.ItemExtent = v

	return l
}

func (l *ListView) isWidget()                {}
func (l *ListView) widgetChildren() []Widget { return l.items }

func (l *ListView) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		ItemExtent: l.ItemExtent,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       l.id,
		Key:      l.key,
		Type:     fugov1.WidgetType_LISTVIEW,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
