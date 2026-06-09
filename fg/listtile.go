package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ListTileWidget is a Material 3 single fixed-height row with a title, optional
// subtitle, and optional leading/trailing icons. Build one with ListTile.
type ListTileWidget struct {
	handler      func(Event)
	title        string
	subtitle     string
	leadingIcon  string
	trailingIcon string
	baseWidget
}

// ListTile creates a list tile with the given title.
func ListTile(title string) *ListTileWidget {
	return &ListTileWidget{title: title}
}

// Subtitle sets the secondary text and returns the widget for chaining.
func (l *ListTileWidget) Subtitle(text string) *ListTileWidget {
	l.subtitle = text

	return l
}

// Leading sets the leading icon by name (see Icon names) and returns the widget for chaining.
func (l *ListTileWidget) Leading(icon string) *ListTileWidget {
	l.leadingIcon = icon

	return l
}

// Trailing sets the trailing icon by name (see Icon names) and returns the widget for chaining.
func (l *ListTileWidget) Trailing(icon string) *ListTileWidget {
	l.trailingIcon = icon

	return l
}

// OnTap registers the tap handler and returns the widget for chaining.
func (l *ListTileWidget) OnTap(handler func(Event)) *ListTileWidget {
	l.handler = handler

	return l
}

func (l *ListTileWidget) isWidget() {}

func (l *ListTileWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnTap handler has been registered.
func (l *ListTileWidget) HasHandler() bool { return l.handler != nil }

// Handle dispatches event to the registered OnTap handler, if any.
func (l *ListTileWidget) Handle(event Event) {
	if l.handler != nil {
		l.handler(event)
	}
}

func (l *ListTileWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	l.id = *counter

	props, _ := proto.Marshal(&fugov1.ListTileProps{
		Title:        l.title,
		Subtitle:     l.subtitle,
		LeadingIcon:  l.leadingIcon,
		TrailingIcon: l.trailingIcon,
	})

	return []*fugov1.WidgetNode{{
		Id:    l.id,
		Key:   l.key,
		Type:  fugov1.WidgetType_LISTTILE,
		Props: props,
	}}
}
