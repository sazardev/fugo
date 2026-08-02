package fg

import (
	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// DismissDirection restricts which swipe directions dismiss a Dismissible.
type DismissDirection int32

// DismissDirection values, matching DismissibleProps.direction on the wire.
const (
	DismissHorizontal DismissDirection = iota // either horizontal direction (default)
	DismissStartToEnd
	DismissEndToStart
	DismissVertical // either vertical direction
	DismissUp
	DismissDown
)

// DismissibleWidget wraps a single child in a swipe-to-dismiss gesture. Build
// one with Dismissible.
type DismissibleWidget struct {
	child      Widget
	onDismiss  func(Event)
	icon       string
	bgColor    style.Color
	direction  DismissDirection
	bgColorSet bool
	baseWidget
}

// Dismissible wraps child in a swipe-to-dismiss gesture.
func Dismissible(child Widget) *DismissibleWidget {
	return &DismissibleWidget{child: child}
}

// Direction restricts which swipe directions dismiss the widget (default
// DismissHorizontal) and returns the widget for chaining.
func (d *DismissibleWidget) Direction(v DismissDirection) *DismissibleWidget {
	d.direction = v

	return d
}

// BgColor sets the color revealed behind the child while swiping (otherwise
// the theme's error container color is used) and returns the widget for chaining.
func (d *DismissibleWidget) BgColor(c style.Color) *DismissibleWidget {
	d.bgColor = c
	d.bgColorSet = true

	return d
}

// Icon sets the icon shown in the revealed background by name (empty = a
// default delete icon) and returns the widget for chaining.
func (d *DismissibleWidget) Icon(name string) *DismissibleWidget {
	d.icon = name

	return d
}

// OnDismissed registers the handler invoked once the child has been swiped
// away; Event.Data carries the dismiss direction name (e.g. "startToEnd").
// Returns the widget for chaining.
func (d *DismissibleWidget) OnDismissed(handler func(Event)) *DismissibleWidget {
	d.onDismiss = handler

	return d
}

func (d *DismissibleWidget) isWidget() {}

func (d *DismissibleWidget) widgetChildren() []Widget {
	if d.child != nil {
		return []Widget{d.child}
	}

	return nil
}

// HasHandler reports whether an OnDismissed handler has been registered.
func (d *DismissibleWidget) HasHandler() bool { return d.onDismiss != nil }

// Handle dispatches event to the registered OnDismissed handler, if any.
func (d *DismissibleWidget) Handle(event Event) {
	if d.onDismiss != nil {
		d.onDismiss(event)
	}
}

func (d *DismissibleWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	d.id = *counter

	childIDs, allNodes := walkChildren(d.widgetChildren(), counter)

	bgColor := ""
	if d.bgColorSet {
		bgColor = d.bgColor.String()
	}

	props, _ := proto.Marshal(&fugov1.DismissibleProps{
		Direction:       int32(d.direction),
		BackgroundColor: bgColor,
		Icon:            d.icon,
	})

	self := &fugov1.WidgetNode{
		Id:       d.id,
		Key:      d.key,
		Type:     fugov1.WidgetType_DISMISSIBLE,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
