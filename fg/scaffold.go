package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ScaffoldWidget is the Material 3 page layout: an optional app bar, a body,
// and an optional floating action button. Build one with Scaffold.
type ScaffoldWidget struct {
	body      Widget
	appBar    Widget
	fab       Widget
	drawer    Widget
	bottomBar Widget
	baseWidget
}

// Scaffold creates a Material scaffold whose content is body.
func Scaffold(body Widget) *ScaffoldWidget {
	return &ScaffoldWidget{body: body}
}

// AppBar sets the top app bar (build one with fg.AppBar) and returns the widget for chaining.
func (s *ScaffoldWidget) AppBar(bar *AppBarWidget) *ScaffoldWidget {
	s.appBar = bar

	return s
}

// FAB sets the floating action button (typically a FloatingActionButton) and returns the widget for chaining.
func (s *ScaffoldWidget) FAB(w Widget) *ScaffoldWidget {
	s.fab = w

	return s
}

// Drawer sets a slide-in side panel (e.g. a Column of ListTiles). When an app
// bar is present, a menu button that opens the drawer appears automatically.
// Returns the widget for chaining.
func (s *ScaffoldWidget) Drawer(w Widget) *ScaffoldWidget {
	s.drawer = w

	return s
}

// BottomBar sets the bottom navigation bar (build one with fg.NavigationBar) and returns the widget for chaining.
func (s *ScaffoldWidget) BottomBar(w Widget) *ScaffoldWidget {
	s.bottomBar = w

	return s
}

func (s *ScaffoldWidget) isWidget() {}

// widgetChildren returns body, then the app bar, then the FAB — the order the
// client relies on (guided by the has_app_bar / has_fab flags) to tell them
// apart.
func (s *ScaffoldWidget) widgetChildren() []Widget {
	var children []Widget
	if s.body != nil {
		children = append(children, s.body)
	}

	if s.appBar != nil {
		children = append(children, s.appBar)
	}

	if s.fab != nil {
		children = append(children, s.fab)
	}

	if s.drawer != nil {
		children = append(children, s.drawer)
	}

	if s.bottomBar != nil {
		children = append(children, s.bottomBar)
	}

	return children
}

func (s *ScaffoldWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	childIDs, allNodes := walkChildren(s.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.ScaffoldProps{
		HasAppBar:    s.appBar != nil,
		HasFab:       s.fab != nil,
		HasDrawer:    s.drawer != nil,
		HasBottomBar: s.bottomBar != nil,
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
