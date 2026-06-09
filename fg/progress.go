package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ProgressWidget is a Material 3 progress indicator. Build one with
// ProgressCircular or ProgressLinear. It is indeterminate (animating) until a
// Value in 0..1 is set.
type ProgressWidget struct {
	value  float64
	linear bool
	baseWidget
}

// ProgressCircular creates an indeterminate circular progress indicator.
func ProgressCircular() *ProgressWidget {
	return &ProgressWidget{value: -1}
}

// ProgressLinear creates an indeterminate linear progress indicator.
func ProgressLinear() *ProgressWidget {
	return &ProgressWidget{value: -1, linear: true}
}

// Value sets a determinate progress fraction in 0..1 and returns the widget for chaining.
func (p *ProgressWidget) Value(v float64) *ProgressWidget {
	p.value = v

	return p
}

func (p *ProgressWidget) isWidget() {}

func (p *ProgressWidget) widgetChildren() []Widget { return nil }

func (p *ProgressWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	p.id = *counter

	props, _ := proto.Marshal(&fugov1.ProgressProps{
		Linear: p.linear,
		Value:  p.value,
	})

	return []*fugov1.WidgetNode{{
		Id:    p.id,
		Key:   p.key,
		Type:  fugov1.WidgetType_PROGRESS,
		Props: props,
	}}
}
