package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// SegmentedButtonWidget is a Material 3 segmented button (a small single-select
// control). Build one with SegmentedButton, add options with Item, set the
// selected value, and read changes with OnChange (the event data is the newly
// selected value).
type SegmentedButtonWidget struct {
	handler  func(Event)
	values   []string
	labels   []string
	selected string
	baseWidget
}

// SegmentedButton creates an empty segmented button; add options with Item.
func SegmentedButton() *SegmentedButtonWidget {
	return &SegmentedButtonWidget{}
}

// Item appends an option with the given value and label, and returns the widget for chaining.
func (s *SegmentedButtonWidget) Item(value, label string) *SegmentedButtonWidget {
	s.values = append(s.values, value)
	s.labels = append(s.labels, label)

	return s
}

// Selected sets the chosen value and returns the widget for chaining.
func (s *SegmentedButtonWidget) Selected(value string) *SegmentedButtonWidget {
	s.selected = value

	return s
}

// OnChange registers the handler invoked when the selection changes; the event
// data is the newly selected value. Returns the widget for chaining.
func (s *SegmentedButtonWidget) OnChange(handler func(Event)) *SegmentedButtonWidget {
	s.handler = handler

	return s
}

func (s *SegmentedButtonWidget) isWidget() {}

func (s *SegmentedButtonWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler is registered.
func (s *SegmentedButtonWidget) HasHandler() bool { return s.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (s *SegmentedButtonWidget) Handle(event Event) {
	if s.handler != nil {
		s.handler(event)
	}
}

func (s *SegmentedButtonWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	props, _ := proto.Marshal(&fugov1.SegmentedButtonProps{
		Values:   s.values,
		Labels:   s.labels,
		Selected: s.selected,
	})

	return []*fugov1.WidgetNode{{
		Id:    s.id,
		Key:   s.key,
		Type:  fugov1.WidgetType_SEGMENTEDBUTTON,
		Props: props,
	}}
}
