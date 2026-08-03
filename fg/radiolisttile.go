package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// RadioListTileWidget is a full-width Material row combining a title/subtitle
// with a trailing radio button, identified by its value and grouped with
// peers sharing the same group value. Build one with RadioListTile.
type RadioListTileWidget struct {
	handler    func(Event)
	title      string
	subtitle   string
	value      string
	groupValue string
	icon       string
	baseWidget
}

// RadioListTile creates a radio list tile carrying value and showing title.
func RadioListTile(value, title string) *RadioListTileWidget {
	return &RadioListTileWidget{value: value, title: title}
}

// Subtitle sets the secondary text and returns the widget for chaining.
func (w *RadioListTileWidget) Subtitle(text string) *RadioListTileWidget {
	w.subtitle = text

	return w
}

// Group sets the shared group value that links mutually-exclusive radio list
// tiles and returns the widget for chaining.
func (w *RadioListTileWidget) Group(groupValue string) *RadioListTileWidget {
	w.groupValue = groupValue

	return w
}

// Icon sets an optional leading icon by name (see Icon names) and returns the widget for chaining.
func (w *RadioListTileWidget) Icon(name string) *RadioListTileWidget {
	w.icon = name

	return w
}

// OnChange registers the handler invoked when this radio list tile is
// selected and returns the widget for chaining.
func (w *RadioListTileWidget) OnChange(handler func(Event)) *RadioListTileWidget {
	w.handler = handler

	return w
}

func (w *RadioListTileWidget) isWidget()                {}
func (w *RadioListTileWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (w *RadioListTileWidget) HasHandler() bool { return w.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (w *RadioListTileWidget) Handle(event Event) {
	if w.handler != nil {
		w.handler(event)
	}
}

func (w *RadioListTileWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	w.id = *counter

	props, err := proto.Marshal(&fugov1.RadioListTileProps{
		Title:         w.title,
		Subtitle:      w.subtitle,
		Value:         w.value,
		GroupValue:    w.groupValue,
		SecondaryIcon: w.icon,
	})
	if err != nil {
		flog.Errorf("marshal RadioListTileProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    w.id,
		Key:   w.key,
		Type:  fugov1.WidgetType_RADIOLISTTILE,
		Props: props,
	}}
}
