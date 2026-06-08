package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type Slider struct {
	handler func(Event)
	Value   float64
	Min     float64
	Max     float64
	baseWidget
}

func Slider() *Slider {
	return &Slider{
		Min: 0,
		Max: 100,
	}
}

func (s *Slider) OnChange(handler func(Event)) *Slider {
	s.handler = handler

	return s
}

func (s *Slider) SetValue(v float64) *Slider {
	s.Value = v

	return s
}

func (s *Slider) SetMin(v float64) *Slider {
	s.Min = v

	return s
}

func (s *Slider) SetMax(v float64) *Slider {
	s.Max = v

	return s
}

func (s *Slider) isWidget()                {}
func (s *Slider) widgetChildren() []Widget { return nil }
func (s *Slider) HasHandler() bool         { return s.handler != nil }

func (s *Slider) Handle(event Event) {
	if s.handler != nil {
		s.handler(event)
	}
}

func (s *Slider) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
