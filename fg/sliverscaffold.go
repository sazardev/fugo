package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// SliverScaffoldWidget is a scrollable screen with a collapsing app bar (a
// CustomScrollView + SliverAppBar + SliverList under the hood). Build one
// with SliverScaffold.
type SliverScaffoldWidget struct {
	title          string
	background     Widget
	body           []Widget
	expandedHeight float64
	pinned         bool
	floating       bool
	baseWidget
}

// SliverScaffold creates a collapsing-header screen titled title, with body
// as the scrollable content below the header. Pinned by default.
func SliverScaffold(title string, body ...Widget) *SliverScaffoldWidget {
	return &SliverScaffoldWidget{title: title, body: body, pinned: true}
}

// ExpandedHeight sets the header's expanded height in logical pixels (0, the
// default, is a plain non-collapsing app bar) and returns the widget for chaining.
func (s *SliverScaffoldWidget) ExpandedHeight(h float64) *SliverScaffoldWidget {
	s.expandedHeight = h

	return s
}

// FlexibleBackground sets a widget rendered behind the title inside the
// collapsing header's flexible space (e.g. an Image) and returns the widget
// for chaining.
func (s *SliverScaffoldWidget) FlexibleBackground(w Widget) *SliverScaffoldWidget {
	s.background = w

	return s
}

// Pinned sets whether the app bar stays visible (collapsed) once scrolled
// past; true by default. Returns the widget for chaining.
func (s *SliverScaffoldWidget) Pinned(v bool) *SliverScaffoldWidget {
	s.pinned = v

	return s
}

// Floating sets whether the app bar reappears as soon as the user scrolls up,
// even mid-list. Returns the widget for chaining.
func (s *SliverScaffoldWidget) Floating(v bool) *SliverScaffoldWidget {
	s.floating = v

	return s
}

func (s *SliverScaffoldWidget) isWidget() {}

func (s *SliverScaffoldWidget) widgetChildren() []Widget {
	if s.background != nil {
		return append([]Widget{s.background}, s.body...)
	}

	return s.body
}

func (s *SliverScaffoldWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	childIDs, allNodes := walkChildren(s.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.SliverScaffoldProps{
		Title:                 s.title,
		ExpandedHeight:        s.expandedHeight,
		Pinned:                s.pinned,
		Floating:              s.floating,
		HasFlexibleBackground: s.background != nil,
	})

	self := &fugov1.WidgetNode{
		Id:       s.id,
		Key:      s.key,
		Type:     fugov1.WidgetType_SLIVERSCAFFOLD,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
