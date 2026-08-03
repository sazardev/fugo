package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// CheckboxListTileWidget is a full-width Material row combining a title/
// subtitle with a trailing checkbox; tapping anywhere in the row toggles it.
// Build one with CheckboxListTile.
type CheckboxListTileWidget struct {
	handler  func(Event)
	title    string
	subtitle string
	checked  bool
	icon     string
	baseWidget
}

// CheckboxListTile creates a checkbox list tile with the given title.
func CheckboxListTile(title string) *CheckboxListTileWidget {
	return &CheckboxListTileWidget{title: title}
}

// Subtitle sets the secondary text and returns the widget for chaining.
func (w *CheckboxListTileWidget) Subtitle(text string) *CheckboxListTileWidget {
	w.subtitle = text

	return w
}

// Checked sets the initial/current checked state and returns the widget for chaining.
func (w *CheckboxListTileWidget) Checked(v bool) *CheckboxListTileWidget {
	w.checked = v

	return w
}

// Icon sets an optional leading icon by name (see Icon names) and returns the widget for chaining.
func (w *CheckboxListTileWidget) Icon(name string) *CheckboxListTileWidget {
	w.icon = name

	return w
}

// OnChange registers the handler invoked when the checked state toggles and
// returns the widget for chaining. Typical usage mirrors Checkbox:
//
//	w.OnChange(func(e fg.Event) { w.Checked(string(e.Data) == "1"); ctx.Update() })
func (w *CheckboxListTileWidget) OnChange(handler func(Event)) *CheckboxListTileWidget {
	w.handler = handler

	return w
}

func (w *CheckboxListTileWidget) isWidget()                {}
func (w *CheckboxListTileWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (w *CheckboxListTileWidget) HasHandler() bool { return w.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (w *CheckboxListTileWidget) Handle(event Event) {
	if w.handler != nil {
		w.handler(event)
	}
}

func (w *CheckboxListTileWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	w.id = *counter

	props, err := proto.Marshal(&fugov1.CheckboxListTileProps{
		Title:         w.title,
		Subtitle:      w.subtitle,
		Checked:       w.checked,
		SecondaryIcon: w.icon,
	})
	if err != nil {
		flog.Errorf("marshal CheckboxListTileProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    w.id,
		Key:   w.key,
		Type:  fugov1.WidgetType_CHECKBOXLISTTILE,
		Props: props,
	}}
}
