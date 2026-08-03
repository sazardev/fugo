package fg

import (
	"github.com/sazardev/fugo/flog"
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// VerticalDividerWidget is a thin vertical rule used to separate content laid
// out in a Row. Build one with VerticalDivider.
type VerticalDividerWidget struct {
	thickness float64
	color     style.Color
	baseWidget
}

// VerticalDivider creates a 1px vertical divider, colored from the active Theme.
func VerticalDivider() *VerticalDividerWidget {
	return &VerticalDividerWidget{
		thickness: 1,
		color:     active.Colors.Border,
	}
}

// Thickness sets the divider line thickness in logical pixels and returns the widget for chaining.
func (v *VerticalDividerWidget) Thickness(t float64) *VerticalDividerWidget {
	v.thickness = t

	return v
}

// Color sets the divider color and returns the widget for chaining.
func (v *VerticalDividerWidget) Color(c style.Color) *VerticalDividerWidget {
	v.color = c

	return v
}

func (v *VerticalDividerWidget) isWidget()                {}
func (v *VerticalDividerWidget) widgetChildren() []Widget { return nil }

func (v *VerticalDividerWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	v.id = *counter

	props, err := proto.Marshal(&fugov1.VerticalDividerProps{
		Thickness: v.thickness,
		Color:     v.color.String(),
	})
	if err != nil {
		flog.Errorf("marshal VerticalDividerProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    v.id,
		Key:   v.key,
		Type:  fugov1.WidgetType_VERTICALDIVIDER,
		Props: props,
	}}
}
