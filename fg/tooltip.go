package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// TooltipWidget wraps a child with a Material tooltip shown on hover or long
// press. Build one with Tooltip.
type TooltipWidget struct {
	child   Widget
	message string
	baseWidget
}

// Tooltip wraps child with a tooltip showing message.
func Tooltip(message string, child Widget) *TooltipWidget {
	return &TooltipWidget{message: message, child: child}
}

func (t *TooltipWidget) isWidget() {}

func (t *TooltipWidget) widgetChildren() []Widget {
	if t.child != nil {
		return []Widget{t.child}
	}

	return nil
}

func (t *TooltipWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	t.id = *counter

	childIDs, allNodes := walkChildren(t.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.TooltipProps{Message: t.message})

	self := &fugov1.WidgetNode{
		Id:       t.id,
		Key:      t.key,
		Type:     fugov1.WidgetType_TOOLTIP,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
