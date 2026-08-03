package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// ResponsiveWidget picks one of several prebuilt children based on the actual
// available width, decided entirely on the client via a LayoutBuilder — there
// is no round trip to Go, unlike most Fugo state changes. Build one with
// Responsive (the default/smallest case), then add wider cases with At.
type ResponsiveWidget struct {
	breakpoints []float64
	children    []Widget
	baseWidget
}

// Responsive creates a responsive widget whose default (smallest) case is
// base, shown whenever no wider breakpoint registered via At matches.
func Responsive(base Widget) *ResponsiveWidget {
	return &ResponsiveWidget{
		breakpoints: []float64{0},
		children:    []Widget{base},
	}
}

// At adds a case shown when the available width is >= minWidth (and no wider
// registered breakpoint also matches); the widest matching case wins. Returns
// the widget for chaining.
func (r *ResponsiveWidget) At(minWidth float64, child Widget) *ResponsiveWidget {
	r.breakpoints = append(r.breakpoints, minWidth)
	r.children = append(r.children, child)

	return r
}

func (r *ResponsiveWidget) isWidget() {}

func (r *ResponsiveWidget) widgetChildren() []Widget { return r.children }

func (r *ResponsiveWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	r.id = *counter

	childIDs, allNodes := walkChildren(r.widgetChildren(), counter)

	props, _ := proto.Marshal(&fugov1.ResponsiveProps{
		Breakpoints: r.breakpoints,
	})

	self := &fugov1.WidgetNode{
		Id:       r.id,
		Key:      r.key,
		Type:     fugov1.WidgetType_RESPONSIVE,
		Props:    props,
		Children: childIDs,
	}

	return append([]*fugov1.WidgetNode{self}, allNodes...)
}
