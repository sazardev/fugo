package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// IconWidget displays a named icon glyph. Build one with Icon.
type IconWidget struct {
	Name  string
	size  float64
	color style.Color
	baseWidget
}

// Icon creates an icon widget for the named glyph, colored from the active Theme.
func Icon(name string) *IconWidget {
	return &IconWidget{
		Name:  name,
		size:  24,
		color: active.Colors.OnSurface,
	}
}

// Size sets the icon size in logical pixels and returns the widget for chaining.
func (i *IconWidget) Size(v float64) *IconWidget {
	i.size = v

	return i
}

// Color sets the icon color and returns the widget for chaining.
func (i *IconWidget) Color(c style.Color) *IconWidget {
	i.color = c

	return i
}

func (i *IconWidget) isWidget()                {}
func (i *IconWidget) widgetChildren() []Widget { return nil }

func (i *IconWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	i.id = *counter

	props, _ := proto.Marshal(&fugov1.IconProps{
		Name:  i.Name,
		Size:  i.size,
		Color: i.color.String(),
	})

	return []*fugov1.WidgetNode{{
		Id:    i.id,
		Key:   i.key,
		Type:  fugov1.WidgetType_ICON,
		Props: props,
	}}
}
