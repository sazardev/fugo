package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// SwitchListTileWidget is a full-width Material row combining a title/
// subtitle with a trailing on/off switch; tapping anywhere in the row toggles
// it. Build one with SwitchListTile.
type SwitchListTileWidget struct {
	handler  func(Event)
	title    string
	subtitle string
	value    bool
	icon     string
	baseWidget
}

// SwitchListTile creates a switch list tile with the given title.
func SwitchListTile(title string) *SwitchListTileWidget {
	return &SwitchListTileWidget{title: title}
}

// Subtitle sets the secondary text and returns the widget for chaining.
func (w *SwitchListTileWidget) Subtitle(text string) *SwitchListTileWidget {
	w.subtitle = text

	return w
}

// Value sets the on/off state and returns the widget for chaining.
func (w *SwitchListTileWidget) Value(v bool) *SwitchListTileWidget {
	w.value = v

	return w
}

// Icon sets an optional leading icon by name (see Icon names) and returns the widget for chaining.
func (w *SwitchListTileWidget) Icon(name string) *SwitchListTileWidget {
	w.icon = name

	return w
}

// OnChange registers the handler invoked when the switch is toggled and
// returns the widget for chaining. Typical usage mirrors Switch:
//
//	w.OnChange(func(e fg.Event) { w.Value(string(e.Data) == "1"); ctx.Update() })
func (w *SwitchListTileWidget) OnChange(handler func(Event)) *SwitchListTileWidget {
	w.handler = handler

	return w
}

func (w *SwitchListTileWidget) isWidget()                {}
func (w *SwitchListTileWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (w *SwitchListTileWidget) HasHandler() bool { return w.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (w *SwitchListTileWidget) Handle(event Event) {
	if w.handler != nil {
		w.handler(event)
	}
}

func (w *SwitchListTileWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	w.id = *counter

	props, err := proto.Marshal(&fugov1.SwitchListTileProps{
		Title:         w.title,
		Subtitle:      w.subtitle,
		Value:         w.value,
		SecondaryIcon: w.icon,
	})
	if err != nil {
		flog.Errorf("marshal SwitchListTileProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    w.id,
		Key:   w.key,
		Type:  fugov1.WidgetType_SWITCHLISTTILE,
		Props: props,
	}}
}
