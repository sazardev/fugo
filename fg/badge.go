package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// BadgeWidget overlays a small Material badge on a child (e.g. a notification
// count on an icon). Build one with Badge; without a Label it shows a dot.
type BadgeWidget struct {
	child Widget
	label string
	baseWidget
}

// Badge wraps child with a badge.
func Badge(child Widget) *BadgeWidget {
	return &BadgeWidget{child: child}
}

// Label sets the badge text (e.g. a count) and returns the widget for chaining.
func (b *BadgeWidget) Label(text string) *BadgeWidget {
	b.label = text

	return b
}

func (b *BadgeWidget) isWidget() {}

func (b *BadgeWidget) widgetChildren() []Widget {
	if b.child != nil {
		return []Widget{b.child}
	}

	return nil
}

func (b *BadgeWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	b.id = *counter

	childIDs, allNodes := walkChildren(b.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.BadgeProps{Label: b.label})

	self := &fugov1.WidgetNode{
		Id:       b.id,
		Key:      b.key,
		Type:     fugov1.WidgetType_BADGE,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
