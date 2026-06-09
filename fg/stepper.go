package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// StepperWidget is a Material vertical stepper (a step-by-step wizard). Build
// one with Stepper and add steps with Step. Read header taps with OnStep (the
// event data is the tapped step index) and drive the current step with Active.
type StepperWidget struct {
	handler  func(Event)
	titles   []string
	contents []Widget
	active   int
	baseWidget
}

// Stepper creates an empty stepper; add steps with Step.
func Stepper() *StepperWidget {
	return &StepperWidget{}
}

// Step appends a step with the given title and content, and returns the widget for chaining.
func (s *StepperWidget) Step(title string, content Widget) *StepperWidget {
	s.titles = append(s.titles, title)
	s.contents = append(s.contents, content)

	return s
}

// Active sets the current step index and returns the widget for chaining.
func (s *StepperWidget) Active(index int) *StepperWidget {
	s.active = index

	return s
}

// OnStep registers the handler invoked when a step header is tapped; the event
// data is the tapped step index. Returns the widget for chaining.
func (s *StepperWidget) OnStep(handler func(Event)) *StepperWidget {
	s.handler = handler

	return s
}

func (s *StepperWidget) isWidget()                {}
func (s *StepperWidget) widgetChildren() []Widget { return s.contents }

// HasHandler reports whether an OnStep handler is registered.
func (s *StepperWidget) HasHandler() bool { return s.handler != nil }

// Handle dispatches event to the registered OnStep handler, if any.
func (s *StepperWidget) Handle(event Event) {
	if s.handler != nil {
		s.handler(event)
	}
}

func (s *StepperWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	s.id = *counter

	ids, nodes := walkChildren(s.contents, counter)
	props, _ := proto.Marshal(&fugov1.StepperProps{
		Titles: s.titles,
		Active: int32(s.active), //nolint:gosec // a small step index
	})

	return selfNode(s.id, s.key, fugov1.WidgetType_STEPPER, props, ids, nodes)
}
