package fg

import (
	"log"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

type RouterWidget struct {
	routes        map[string]func() Widget
	current       string
	history       []string
	currentWidget Widget
	baseWidget
}

func Router(routes map[string]func() Widget, initialRoute string) *RouterWidget {
	return &RouterWidget{
		routes:  routes,
		current: initialRoute,
	}
}

func (r *RouterWidget) NavigateTo(route string) bool {
	if _, ok := r.routes[route]; !ok {
		log.Printf("[router] route not found: %s", route)

		return false
	}
	if r.current != "" {
		r.history = append(r.history, r.current)
	}
	r.current = route
	r.currentWidget = nil
	log.Printf("[router] navigate to: %s (history: %d)", route, len(r.history))

	return true
}

func (r *RouterWidget) GoBack() bool {
	if len(r.history) == 0 {
		log.Println("[router] goback: no history")

		return false
	}
	r.current = r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
	r.currentWidget = nil
	log.Printf("[router] goback to: %s (history: %d)", r.current, len(r.history))

	return true
}

func (r *RouterWidget) CurrentRoute() string {
	return r.current
}

func (r *RouterWidget) isWidget() {}

func (r *RouterWidget) widgetChildren() []Widget {
	if r.currentWidget != nil {
		return []Widget{r.currentWidget}
	}

	return nil
}

func (r *RouterWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	r.id = *counter

	builder, ok := r.routes[r.current]
	if !ok {
		r.currentWidget = nil

		return []*fugov1.WidgetNode{{
			Id:   r.id,
			Key:  "router",
			Type: fugov1.WidgetType_CONTAINER,
		}}
	}

	if r.currentWidget == nil {
		r.currentWidget = builder()
	}
	childNodes := r.currentWidget.walkNodes(counter)

	var childIDs []uint32
	if len(childNodes) > 0 {
		childIDs = append(childIDs, childNodes[0].GetId())
	}

	return append([]*fugov1.WidgetNode{{
		Id:       r.id,
		Key:      "router",
		Type:     fugov1.WidgetType_CONTAINER,
		Children: childIDs,
	}}, childNodes...)
}
