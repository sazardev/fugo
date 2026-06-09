package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ExpansionTileWidget is a Material expand/collapse tile (an accordion row).
// Build one with ExpansionTile and set its expandable content with Children.
// Expanding and collapsing are handled on the client.
type ExpansionTileWidget struct {
	items       []Widget
	title       string
	subtitle    string
	leadingIcon string
	expanded    bool
	baseWidget
}

// ExpansionTile creates a collapsible tile with the given title.
func ExpansionTile(title string) *ExpansionTileWidget {
	return &ExpansionTileWidget{title: title}
}

// Children sets the content revealed when the tile is expanded, and returns the widget for chaining.
func (e *ExpansionTileWidget) Children(ws ...Widget) *ExpansionTileWidget {
	e.items = ws

	return e
}

// Subtitle sets the secondary line and returns the widget for chaining.
func (e *ExpansionTileWidget) Subtitle(text string) *ExpansionTileWidget {
	e.subtitle = text

	return e
}

// Leading sets the leading icon by name (see fg.Icons) and returns the widget for chaining.
func (e *ExpansionTileWidget) Leading(icon string) *ExpansionTileWidget {
	e.leadingIcon = icon

	return e
}

// InitiallyExpanded controls whether the tile starts open and returns the widget for chaining.
func (e *ExpansionTileWidget) InitiallyExpanded(v bool) *ExpansionTileWidget {
	e.expanded = v

	return e
}

func (e *ExpansionTileWidget) isWidget()                {}
func (e *ExpansionTileWidget) widgetChildren() []Widget { return e.items }

func (e *ExpansionTileWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	e.id = *counter
	ids, nodes := walkChildren(e.widgetChildren(), counter)
	props, _ := proto.Marshal(&fugov1.ExpansionTileProps{
		Title:             e.title,
		Subtitle:          e.subtitle,
		LeadingIcon:       e.leadingIcon,
		InitiallyExpanded: e.expanded,
	})

	return selfNode(e.id, e.key, fugov1.WidgetType_EXPANSIONTILE, props, ids, nodes)
}
