package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// RadioWidget is a single radio button identified by its Value and grouped with
// peers sharing the same GroupValue. Build one with Radio.
type RadioWidget struct {
	handler    func(Event)
	Value      string
	GroupValue string
	Label      string
	baseWidget
}

// Radio creates a radio button carrying value and showing label.
func Radio(value, label string) *RadioWidget {
	return &RadioWidget{Value: value, Label: label}
}

// Group sets the shared group value that links mutually-exclusive radios and
// returns the widget for chaining.
func (r *RadioWidget) Group(groupValue string) *RadioWidget {
	r.GroupValue = groupValue

	return r
}

// OnChange registers the handler invoked when this radio is selected and
// returns the widget for chaining.
func (r *RadioWidget) OnChange(handler func(Event)) *RadioWidget {
	r.handler = handler

	return r
}

func (r *RadioWidget) isWidget()                {}
func (r *RadioWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (r *RadioWidget) HasHandler() bool { return r.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (r *RadioWidget) Handle(event Event) {
	if r.handler != nil {
		r.handler(event)
	}
}

func (r *RadioWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	r.id = *counter

	props, _ := proto.Marshal(&fugov1.RadioProps{
		Value:      r.Value,
		GroupValue: r.GroupValue,
		Label:      r.Label,
	})

	return []*fugov1.WidgetNode{{
		Id:    r.id,
		Key:   r.key,
		Type:  fugov1.WidgetType_RADIO,
		Props: props,
	}}
}
