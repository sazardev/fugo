package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// TabsWidget is a Material tab bar with a swipeable view per tab. Build one with
// Tabs and add tabs with Tab(label, content). Switching tabs is handled on the
// client, so it needs no round-trip to Go.
type TabsWidget struct {
	labels  []string
	views   []Widget
	initial int
	baseWidget
}

// Tabs creates an empty tab strip; add tabs with Tab.
func Tabs() *TabsWidget {
	return &TabsWidget{}
}

// Tab appends a tab with the given label and its content view, and returns the widget for chaining.
func (t *TabsWidget) Tab(label string, content Widget) *TabsWidget {
	t.labels = append(t.labels, label)
	t.views = append(t.views, content)

	return t
}

// InitialIndex sets the tab shown first and returns the widget for chaining.
func (t *TabsWidget) InitialIndex(i int) *TabsWidget {
	t.initial = i

	return t
}

func (t *TabsWidget) isWidget() {}

func (t *TabsWidget) widgetChildren() []Widget { return t.views }

func (t *TabsWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	childIDs, allNodes := walkChildren(t.views, counter)

	props, _ := proto.Marshal(&fugov1.TabsProps{
		Labels:       t.labels,
		InitialIndex: int32(t.initial), //nolint:gosec // a small tab index
	})

	self := &fugov1.WidgetNode{
		Id:       t.id,
		Key:      t.key,
		Type:     fugov1.WidgetType_TABS,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
