package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type AnimatedOpacityWidget struct {
	child      Widget
	opacity    float64
	durationMs int32
	baseWidget
}

func AnimatedOpacity(child Widget) *AnimatedOpacityWidget {
	return &AnimatedOpacityWidget{
		child:      child,
		opacity:    1,
		durationMs: 200,
	}
}

func (o *AnimatedOpacityWidget) Opacity(v float64) *AnimatedOpacityWidget {
	o.opacity = v

	return o
}

func (o *AnimatedOpacityWidget) DurationMs(v int32) *AnimatedOpacityWidget {
	o.durationMs = v

	return o
}

func (o *AnimatedOpacityWidget) isWidget() {}

func (o *AnimatedOpacityWidget) widgetChildren() []Widget {
	if o.child != nil {
		return []Widget{o.child}
	}

	return nil
}

func (o *AnimatedOpacityWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	o.id = *counter

	var childIDs []uint32
	var allNodes []*fugov1.WidgetNode

	for _, child := range o.widgetChildren() {
		subNodes := child.walkNodes(counter)
		if len(subNodes) > 0 {
			childIDs = append(childIDs, subNodes[0].GetId())
			allNodes = append(allNodes, subNodes...)
		}
	}

	props, _ := proto.Marshal(&fugov1.AnimatedOpacityProps{
		Opacity:    o.opacity,
		DurationMs: o.durationMs,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       o.id,
		Key:      o.key,
		Type:     fugov1.WidgetType_ANIMATEDOPACITY,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
