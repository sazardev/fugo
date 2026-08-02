package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// RangeSliderWidget lets the user pick a numeric [start, end] range within a
// bound by dragging two thumbs. Build one with RangeSlider.
type RangeSliderWidget struct {
	handler    func(Event)
	start, end float64
	min, max   float64
	baseWidget
}

// RangeSlider creates a range slider with a default range of 0 to 100.
func RangeSlider() *RangeSliderWidget {
	return &RangeSliderWidget{
		min: 0,
		max: 100,
	}
}

// SetStart sets the current lower thumb value and returns the widget for chaining.
func (r *RangeSliderWidget) SetStart(v float64) *RangeSliderWidget {
	r.start = v

	return r
}

// SetEnd sets the current upper thumb value and returns the widget for chaining.
func (r *RangeSliderWidget) SetEnd(v float64) *RangeSliderWidget {
	r.end = v

	return r
}

// SetMin sets the minimum value of the range and returns the widget for chaining.
func (r *RangeSliderWidget) SetMin(v float64) *RangeSliderWidget {
	r.min = v

	return r
}

// SetMax sets the maximum value of the range and returns the widget for chaining.
func (r *RangeSliderWidget) SetMax(v float64) *RangeSliderWidget {
	r.max = v

	return r
}

// OnChange registers the handler invoked as the range slider values change
// and returns the widget for chaining. The event's Data carries the new
// "start,end" values as two decimals separated by a comma, each formatted
// to 2 decimal places (e.g. "20.50,80.00", mirroring the Slider widget's
// value.toStringAsFixed(2) on the Dart side) — split on "," and parse both
// sides with strconv.ParseFloat.
func (r *RangeSliderWidget) OnChange(handler func(Event)) *RangeSliderWidget {
	r.handler = handler

	return r
}

func (r *RangeSliderWidget) isWidget()                {}
func (r *RangeSliderWidget) widgetChildren() []Widget { return nil }

// HasHandler reports whether an OnChange handler has been registered.
func (r *RangeSliderWidget) HasHandler() bool { return r.handler != nil }

// Handle dispatches event to the registered OnChange handler, if any.
func (r *RangeSliderWidget) Handle(event Event) {
	if r.handler != nil {
		r.handler(event)
	}
}

func (r *RangeSliderWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	r.id = *counter

	props, err := proto.Marshal(&fugov1.RangeSliderProps{
		Start: r.start,
		End:   r.end,
		Min:   r.min,
		Max:   r.max,
	})
	if err != nil {
		flog.Errorf("marshal RangeSliderProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    r.id,
		Key:   r.key,
		Type:  fugov1.WidgetType_RANGESLIDER,
		Props: props,
	}}
}
