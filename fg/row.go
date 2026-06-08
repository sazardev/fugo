package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// RowWidget arranges its children horizontally with configurable axis alignment. Build one with Row.
type RowWidget struct {
	items        []Widget
	mainAxisSize fugov1.MainAxisSize
	mainAlign    fugov1.MainAxisAlignment
	crossAlign   fugov1.CrossAxisAlignment
	baseWidget
}

// Row creates a horizontal layout containing the given items.
func Row(items ...Widget) *RowWidget {
	return &RowWidget{
		items:      items,
		crossAlign: fugov1.CrossAxisAlignment_CROSS_CENTER,
	}
}

// MainAlign sets how children are distributed along the horizontal axis and returns the widget for chaining.
func (r *RowWidget) MainAlign(v fugov1.MainAxisAlignment) *RowWidget {
	r.mainAlign = v

	return r
}

// CrossAlign sets how children are aligned along the vertical axis and returns the widget for chaining.
func (r *RowWidget) CrossAlign(v fugov1.CrossAxisAlignment) *RowWidget {
	r.crossAlign = v

	return r
}

// MainAxisSize sets whether the row shrinks to fit its children or fills the horizontal axis, and returns the widget for chaining.
func (r *RowWidget) MainAxisSize(v fugov1.MainAxisSize) *RowWidget {
	r.mainAxisSize = v

	return r
}

func (r *RowWidget) isWidget()                {}
func (r *RowWidget) widgetChildren() []Widget { return r.items }

func (r *RowWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		MainAxisSize:   r.mainAxisSize,
		MainAlignment:  r.mainAlign,
		CrossAlignment: r.crossAlign,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       r.id,
		Key:      r.key,
		Type:     fugov1.WidgetType_ROW,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
