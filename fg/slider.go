package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type SliderWidget struct {
	handler func(Event)
	Value   float64
	Min     float64
	Max     float64
	baseWidget
}

func Slider() *SliderWidget {
	return &SliderWidget{
		Min: 0,
		Max: 100,
	}
}

func (s *SliderWidget) OnChange(handler func(Event)) *SliderWidget {
	s.handler = handler

	return s
}

func (s *SliderWidget) SetValue(v float64) *SliderWidget {
	s.Value = v

	return s
}

func (s *SliderWidget) SetMin(v float64) *SliderWidget {
	s.Min = v

	return s
}

func (s *SliderWidget) SetMax(v float64) *SliderWidget {
	s.Max = v

	return s
}

func (s *SliderWidget) isWidget()                {}
func (s *SliderWidget) widgetChildren() []Widget { return nil }
func (s *SliderWidget) HasHandler() bool         { return s.handler != nil }

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
