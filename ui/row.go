package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Row struct {
	items        []Widget
	MainAxisSize fugov1.MainAxisSize
	MainAlign    fugov1.MainAxisAlignment
	CrossAlign   fugov1.CrossAxisAlignment
	baseWidget
}

func NewRow(items ...Widget) *Row {
	return &Row{
		items:      items,
		CrossAlign: fugov1.CrossAxisAlignment_CROSS_CENTER,
	}
}

func (r *Row) WithMainAlign(v fugov1.MainAxisAlignment) *Row {
	r.MainAlign = v

	return r
}

func (r *Row) WithCrossAlign(v fugov1.CrossAxisAlignment) *Row {
	r.CrossAlign = v

	return r
}

func (r *Row) WithMainAxisSize(v fugov1.MainAxisSize) *Row {
	r.MainAxisSize = v

	return r
}

func (r *Row) isWidget()                {}
func (r *Row) widgetChildren() []Widget { return r.items }

func (r *Row) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	r.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range r.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.RowProps{
		MainAxisSize:   r.MainAxisSize,
		MainAlignment:  r.MainAlign,
		CrossAlignment: r.CrossAlign,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       r.id,
		Key:      r.key,
		Type:     fugov1.WidgetType_ROW,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
