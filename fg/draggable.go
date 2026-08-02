package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// DraggableWidget wraps a single child that the user can drag onto a
// DragTarget elsewhere in the tree. Build one with Draggable.
type DraggableWidget struct {
	child Widget
	data  string
	baseWidget
}

// Draggable wraps child in a drag gesture.
func Draggable(child Widget) *DraggableWidget {
	return &DraggableWidget{child: child}
}

// Data sets the opaque payload delivered to whatever DragTarget accepts a
// drop of this widget, and returns the widget for chaining.
func (d *DraggableWidget) Data(v string) *DraggableWidget {
	d.data = v

	return d
}

func (d *DraggableWidget) isWidget() {}

func (d *DraggableWidget) widgetChildren() []Widget {
	if d.child != nil {
		return []Widget{d.child}
	}

	return nil
}

func (d *DraggableWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	childIDs, allNodes := walkChildren(d.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.DraggableProps{Data: d.data})

	self := &fugov1.WidgetNode{
		Id:       d.id,
		Key:      d.key,
		Type:     fugov1.WidgetType_DRAGGABLE,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}

// DragTargetWidget wraps a single child and accepts drops from a Draggable.
// Build one with DragTarget.
type DragTargetWidget struct {
	child    Widget
	onAccept func(Event)
	baseWidget
}

// DragTarget wraps child, accepting drops from any Draggable.
func DragTarget(child Widget) *DragTargetWidget {
	return &DragTargetWidget{child: child}
}

// OnAccept registers the handler invoked when a Draggable is dropped on this
// target; Event.Data carries the dropped Draggable's Data. Returns the widget
// for chaining.
func (d *DragTargetWidget) OnAccept(handler func(Event)) *DragTargetWidget {
	d.onAccept = handler

	return d
}

func (d *DragTargetWidget) isWidget() {}

func (d *DragTargetWidget) widgetChildren() []Widget {
	if d.child != nil {
		return []Widget{d.child}
	}

	return nil
}

// HasHandler reports whether an OnAccept handler has been registered.
func (d *DragTargetWidget) HasHandler() bool { return d.onAccept != nil }

// Handle dispatches event to the registered OnAccept handler, if any.
func (d *DragTargetWidget) Handle(event Event) {
	if d.onAccept != nil {
		d.onAccept(event)
	}
}

func (d *DragTargetWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	childIDs, allNodes := walkChildren(d.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.DragTargetProps{})

	self := &fugov1.WidgetNode{
		Id:       d.id,
		Key:      d.key,
		Type:     fugov1.WidgetType_DRAGTARGET,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
