package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AppBarWidget is a Material 3 top app bar. Build one with AppBar and pass it to
// Scaffold().AppBar(...). It carries a title plus an optional leading widget
// (e.g. a menu icon button) and trailing action widgets.
type AppBarWidget struct {
	title       string
	leading     Widget
	actions     []Widget
	bgColor     Color
	centerTitle bool
	bgColorSet  bool
	baseWidget
}

// AppBar creates a top app bar showing title.
func AppBar(title string) *AppBarWidget {
	return &AppBarWidget{title: title}
}

// Leading sets the widget shown before the title — e.g.
// fg.IconButton(fg.Icons.Menu) — and returns the widget for chaining.
func (a *AppBarWidget) Leading(w Widget) *AppBarWidget {
	a.leading = w

	return a
}

// Actions sets the trailing action widgets (typically icon buttons) and returns the widget for chaining.
func (a *AppBarWidget) Actions(ws ...Widget) *AppBarWidget {
	a.actions = ws

	return a
}

// CenterTitle centers the title and returns the widget for chaining.
func (a *AppBarWidget) CenterTitle(v bool) *AppBarWidget {
	a.centerTitle = v

	return a
}

// BgColor overrides the app bar background (otherwise the M3 surface) and returns the widget for chaining.
func (a *AppBarWidget) BgColor(c Color) *AppBarWidget {
	a.bgColor = c
	a.bgColorSet = true

	return a
}

func (a *AppBarWidget) isWidget() {}

// widgetChildren returns the leading widget (when set) followed by the actions.
func (a *AppBarWidget) widgetChildren() []Widget {
	var children []Widget
	if a.leading != nil {
		children = append(children, a.leading)
	}

	return append(children, a.actions...)
}

func (a *AppBarWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	a.id = *counter

	childIDs, allNodes := walkChildren(a.widgetChildren(), counter)

	bgColor := ""
	if a.bgColorSet {
		bgColor = a.bgColor.String()
	}

	props, _ := proto.Marshal(&fugov1.AppBarProps{
		Title:       a.title,
		CenterTitle: a.centerTitle,
		HasLeading:  a.leading != nil,
		BgColor:     bgColor,
	})

	self := &fugov1.WidgetNode{
		Id:       a.id,
		Key:      a.key,
		Type:     fugov1.WidgetType_APPBAR,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
