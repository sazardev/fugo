package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// SliderWidget lets the user pick a numeric value within a range by dragging. Build one with Slider.
type SliderWidget struct {
	handler func(Event)
	Value   float64
	Min     float64
	Max     float64
	baseWidget
}

// Slider creates a slider with a default range of 0 to 100.
func Slider() *SliderWidget {
	return &SliderWidget{
		Min: 0,
		Max: 100,
	}
}

// OnChange registers the handler invoked as the slider value changes and returns the widget for chaining.
func (s *SliderWidget) OnChange(handler func(Event)) *SliderWidget {
	s.handler = handler

	return s
}

// SetValue sets the current value and returns the widget for chaining.
func (s *SliderWidget) SetValue(v float64) *SliderWidget {
	s.Value = v

	return s
}

// SetMin sets the minimum value of the range and returns the widget for chaining.
func (s *SliderWidget) SetMin(v float64) *SliderWidget {
	s.Min = v

	return s
}

// SetMax sets the maximum value of the range and returns the widget for chaining.
func (s *SliderWidget) SetMax(v float64) *SliderWidget {
	s.Max = v

	return s
}

func (s *SliderWidget) isWidget()                {}
func (s *SliderWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (s *SliderWidget) HasHandler() bool { return s.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (s *SliderWidget) Handle(event Event) {
	if s.handler != nil {
		s.handler(event)
	}
}

func (s *SliderWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	props, _ := proto.Marshal(&fugov1.SliderProps{
		Value: s.Value,
		Min:   s.Min,
		Max:   s.Max,
	})

	return []*fugov1.WidgetNode{{
		Id:    s.id,
		Key:   s.key,
		Type:  fugov1.WidgetType_SLIDER,
		Props: props,
	}}
}
