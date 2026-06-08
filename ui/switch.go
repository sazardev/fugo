package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type SwitchWidget struct {
	handler func(Event)
	Value   bool
	baseWidget
}

func NewSwitch() *SwitchWidget {
	return &SwitchWidget{}
}

func (s *SwitchWidget) OnChange(handler func(Event)) *SwitchWidget {
	s.handler = handler

	return s
}

func (s *SwitchWidget) WithValue(v bool) *SwitchWidget {
	s.Value = v

	return s
}

func (s *SwitchWidget) isWidget()                {}
func (s *SwitchWidget) widgetChildren() []Widget { return nil }
func (s *SwitchWidget) HasHandler() bool         { return s.handler != nil }

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
