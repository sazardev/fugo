package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ScrollbarWidget wraps a single scrollable child with an explicit,
// draggable scrollbar. Build one with Scrollbar.
type ScrollbarWidget struct {
	child         Widget
	alwaysVisible bool
	baseWidget
}

// Scrollbar wraps child with an explicit scrollbar.
func Scrollbar(child Widget) *ScrollbarWidget {
	return &ScrollbarWidget{child: child}
}

// AlwaysVisible controls whether the scrollbar thumb stays visible even when
// not scrolling, and returns the widget for chaining.
func (s *ScrollbarWidget) AlwaysVisible(v bool) *ScrollbarWidget {
	s.alwaysVisible = v

	return s
}

func (s *ScrollbarWidget) isWidget() {}

func (s *ScrollbarWidget) widgetChildren() []Widget {
	if s.child != nil {
		return []Widget{s.child}
	}

	return nil
}

func (s *ScrollbarWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	var (
		childIDs []uint32
		allNodes []*fugov1.WidgetNode
	)

	for _, child := range s.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, err := proto.Marshal(&fugov1.ScrollbarProps{AlwaysVisible: s.alwaysVisible})
	if err != nil {
		flog.Errorf("marshal ScrollbarProps: %v", err)
	}

	return append([]*fugov1.WidgetNode{{
		Id:       s.id,
		Key:      s.key,
		Type:     fugov1.WidgetType_SCROLLBAR,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}

// RefreshIndicatorWidget wraps a single scrollable child with Material
// pull-to-refresh. Build one with RefreshIndicator.
//
// Limitation: OnRefresh is fire-and-forget. Go's handler runs synchronously
// on event dispatch, mutates state, and calls ctx.Update() itself; the
// client-side spinner resolves immediately after sending the event and does
// NOT wait for Go's response or for the next render to land. If the refresh
// takes a while to reflect in the UI, the spinner will already be gone.
type RefreshIndicatorWidget struct {
	handler func(Event)
	child   Widget
	baseWidget
}

// RefreshIndicator wraps child with Material pull-to-refresh.
func RefreshIndicator(child Widget) *RefreshIndicatorWidget {
	return &RefreshIndicatorWidget{child: child}
}

// OnRefresh registers the handler invoked when the user pulls to refresh.
// See the type doc for the fire-and-forget caveat.
func (r *RefreshIndicatorWidget) OnRefresh(handler func(Event)) *RefreshIndicatorWidget {
	r.handler = handler

	return r
}

func (r *RefreshIndicatorWidget) isWidget() {}

func (r *RefreshIndicatorWidget) widgetChildren() []Widget {
	if r.child != nil {
		return []Widget{r.child}
	}

	return nil
}

// HasHandler reports whether an OnRefresh handler has been registered.
func (r *RefreshIndicatorWidget) HasHandler() bool { return r.handler != nil }

// Handle dispatches event to the registered OnRefresh handler, if any.
func (r *RefreshIndicatorWidget) Handle(event Event) {
	if r.handler != nil {
		r.handler(event)
	}
}

func (r *RefreshIndicatorWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	r.id = *counter

	var (
		childIDs []uint32
		allNodes []*fugov1.WidgetNode
	)

	for _, child := range r.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, err := proto.Marshal(&fugov1.RefreshIndicatorProps{})
	if err != nil {
		flog.Errorf("marshal RefreshIndicatorProps: %v", err)
	}

	return append([]*fugov1.WidgetNode{{
		Id:       r.id,
		Key:      r.key,
		Type:     fugov1.WidgetType_REFRESHINDICATOR,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
