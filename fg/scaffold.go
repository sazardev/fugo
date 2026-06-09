package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ScaffoldWidget is the Material 3 page layout: an optional app bar, a body,
// and an optional floating action button. Build one with Scaffold.
type ScaffoldWidget struct {
	body        Widget
	fab         Widget
	appBarTitle string
	hasAppBar   bool
	baseWidget
}

// Scaffold creates a Material scaffold whose content is body.
func Scaffold(body Widget) *ScaffoldWidget {
	return &ScaffoldWidget{body: body}
}

// AppBar adds a top app bar with the given title and returns the widget for chaining.
func (s *ScaffoldWidget) AppBar(title string) *ScaffoldWidget {
	s.appBarTitle = title
	s.hasAppBar = true

	return s
}

// FAB sets the floating action button (typically a FloatingActionButton) and returns the widget for chaining.
func (s *ScaffoldWidget) FAB(w Widget) *ScaffoldWidget {
	s.fab = w

	return s
}

func (s *ScaffoldWidget) isWidget() {}

// widgetChildren returns the body first and, when present, the FAB last — the
// order the client relies on to tell them apart.
func (s *ScaffoldWidget) widgetChildren() []Widget {
	var children []Widget
	if s.body != nil {
		children = append(children, s.body)
	}

	if s.fab != nil {
		children = append(children, s.fab)
	}

	return children
}

func (s *ScaffoldWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	childIDs, allNodes := walkChildren(s.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.ScaffoldProps{
		AppBarTitle: s.appBarTitle,
		HasAppBar:   s.hasAppBar,
		HasFab:      s.fab != nil,
	})

	self := &fugov1.WidgetNode{
		Id:       s.id,
		Key:      s.key,
		Type:     fugov1.WidgetType_SCAFFOLD,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
