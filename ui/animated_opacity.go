package ui

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

type AnimatedOpacity struct {
	child      Widget
	Opacity    float64
	DurationMs int32
	baseWidget
}

func NewAnimatedOpacity(child Widget) *AnimatedOpacity {
	return &AnimatedOpacity{
		child:      child,
		Opacity:    1,
		DurationMs: 200,
	}
}

func (o *AnimatedOpacity) WithOpacity(v float64) *AnimatedOpacity {
	o.Opacity = v

	return o
}

func (o *AnimatedOpacity) WithDurationMs(v int32) *AnimatedOpacity {
	o.DurationMs = v

	return o
}

func (o *AnimatedOpacity) isWidget() {}

func (o *AnimatedOpacity) widgetChildren() []Widget {
	if o.child != nil {
		return []Widget{o.child}
	}

	return nil
}

func (o *AnimatedOpacity) walkNodes(counter *uint32) []*fugov1.WidgetNode {
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
		Opacity:    o.Opacity,
		DurationMs: o.DurationMs,
	})

	return append([]*fugov1.WidgetNode{{
		Id:       o.id,
		Key:      o.key,
		Type:     fugov1.WidgetType_ANIMATEDOPACITY,
		Props:    props,
		Children: childIDs,
	}}, allNodes...)
}
