package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// DividerWidget is a thin horizontal rule used to separate content. Build one with Divider.
type DividerWidget struct {
	thickness float64
	color     style.Color
	baseWidget
}

// Divider creates a 1px horizontal divider, colored from the active Theme.
func Divider() *DividerWidget {
	return &DividerWidget{
		thickness: 1,
		color:     active.Colors.Border,
	}
}

// Thickness sets the divider line thickness in logical pixels and returns the widget for chaining.
func (d *DividerWidget) Thickness(v float64) *DividerWidget {
	d.thickness = v

	return d
}

// Color sets the divider color and returns the widget for chaining.
func (d *DividerWidget) Color(c style.Color) *DividerWidget {
	d.color = c

	return d
}

func (d *DividerWidget) isWidget()                {}
func (d *DividerWidget) widgetChildren() []Widget { return nil }

func (d *DividerWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	props, _ := proto.Marshal(&fugov1.DividerProps{
		Thickness: d.thickness,
		Color:     d.color.String(),
	})

	return []*fugov1.WidgetNode{{
		Id:    d.id,
		Key:   d.key,
		Type:  fugov1.WidgetType_DIVIDER,
		Props: props,
	}}
}
