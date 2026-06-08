package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// SwitchWidget is an on/off toggle switch. Build one with Switch.
type SwitchWidget struct {
	handler func(Event)
	Value   bool
	baseWidget
}

// Switch creates a toggle switch that is off by default.
func Switch() *SwitchWidget {
	return &SwitchWidget{}
}

// OnChange registers the handler invoked when the switch is toggled and returns the widget for chaining.
func (s *SwitchWidget) OnChange(handler func(Event)) *SwitchWidget {
	s.handler = handler

	return s
}

// SetValue sets the on/off state and returns the widget for chaining.
func (s *SwitchWidget) SetValue(v bool) *SwitchWidget {
	s.Value = v

	return s
}

func (s *SwitchWidget) isWidget()                {}
func (s *SwitchWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (s *SwitchWidget) HasHandler() bool { return s.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (s *SwitchWidget) Handle(event Event) {
	if s.handler != nil {
		s.handler(event)
	}
}

func (s *SwitchWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	props, _ := proto.Marshal(&fugov1.SwitchProps{
		Value: s.Value,
	})

	return []*fugov1.WidgetNode{{
		Id:    s.id,
		Key:   s.key,
		Type:  fugov1.WidgetType_SWITCH_WIDGET,
		Props: props,
	}}
}
