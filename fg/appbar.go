package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AppBarWidget is a Material 3 top app bar. Build one with AppBar and pass it to
// Scaffold().AppBar(...). It carries a title plus an optional leading widget
// (e.g. a menu icon button), trailing action widgets, and an optional bottom
// widget docked under the title row (a fixed-height strip — e.g. a search
// field or a row of filter chips; not a place to embed a full Tabs, whose
// TabBarView content belongs in the Scaffold body instead).
type AppBarWidget struct {
	title           string
	leading         Widget
	actions         []Widget
	bottom          Widget
	bgColor         Color
	foregroundColor Color
	elevation       float64
	centerTitle     bool
	bgColorSet      bool
	foregroundSet   bool
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

// Bottom sets a fixed-height widget docked under the title row (e.g. a search
// field or a row of filter chips) and returns the widget for chaining.
func (a *AppBarWidget) Bottom(w Widget) *AppBarWidget {
	a.bottom = w

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

// ForegroundColor overrides the title/icon color (otherwise the M3 default)
// and returns the widget for chaining.
func (a *AppBarWidget) ForegroundColor(c Color) *AppBarWidget {
	a.foregroundColor = c
	a.foregroundSet = true

	return a
}

// Elevation sets the shadow depth in logical pixels (0 = the M3 default) and
// returns the widget for chaining.
func (a *AppBarWidget) Elevation(v float64) *AppBarWidget {
	a.elevation = v

	return a
}

func (a *AppBarWidget) isWidget() {}

// widgetChildren returns, in order: the leading widget (when set), the
// actions, and finally the bottom widget (when set) — this fixed order is
// what lets the client tell them apart without a separate count.
func (a *AppBarWidget) widgetChildren() []Widget {
	var children []Widget
	if a.leading != nil {
		children = append(children, a.leading)
	}

	children = append(children, a.actions...)

	if a.bottom != nil {
		children = append(children, a.bottom)
	}

	return children
}

func (a *AppBarWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	a.id = *counter

	childIDs, allNodes := walkChildren(a.widgetChildren(), counter)

	bgColor := ""
	if a.bgColorSet {
		bgColor = a.bgColor.String()
	}

	foregroundColor := ""
	if a.foregroundSet {
		foregroundColor = a.foregroundColor.String()
	}

	props, _ := proto.Marshal(&fugov1.AppBarProps{
		Title:           a.title,
		CenterTitle:     a.centerTitle,
		HasLeading:      a.leading != nil,
		BgColor:         bgColor,
		Elevation:       a.elevation,
		ForegroundColor: foregroundColor,
		HasBottom:       a.bottom != nil,
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
