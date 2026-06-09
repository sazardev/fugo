package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// AvatarWidget is a Material CircleAvatar showing initials or an icon. Build one
// with CircleAvatar.
type AvatarWidget struct {
	text       string
	icon       string
	radius     float64
	bgColor    Color
	bgColorSet bool
	baseWidget
}

// CircleAvatar creates a circular avatar; set its content with Text or Icon.
func CircleAvatar() *AvatarWidget {
	return &AvatarWidget{}
}

// Text sets the initials shown in the avatar and returns the widget for chaining.
func (a *AvatarWidget) Text(s string) *AvatarWidget {
	a.text = s

	return a
}

// Icon sets the icon shown when no Text is set (see fg.Icons) and returns the widget for chaining.
func (a *AvatarWidget) Icon(name string) *AvatarWidget {
	a.icon = name

	return a
}

// Radius sets the avatar radius in logical pixels and returns the widget for chaining.
func (a *AvatarWidget) Radius(r float64) *AvatarWidget {
	a.radius = r

	return a
}

// BgColor sets the avatar background color and returns the widget for chaining.
func (a *AvatarWidget) BgColor(c Color) *AvatarWidget {
	a.bgColor = c
	a.bgColorSet = true

	return a
}

func (a *AvatarWidget) isWidget() {}

func (a *AvatarWidget) widgetChildren() []Widget { return nil }

func (a *AvatarWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	a.id = *counter

	bgColor := ""
	if a.bgColorSet {
		bgColor = a.bgColor.String()
	}

	props, _ := proto.Marshal(&fugov1.AvatarProps{
		Text:    a.text,
		Icon:    a.icon,
		BgColor: bgColor,
		Radius:  a.radius,
	})

	return []*fugov1.WidgetNode{{
		Id:    a.id,
		Key:   a.key,
		Type:  fugov1.WidgetType_AVATAR,
		Props: props,
	}}
}
