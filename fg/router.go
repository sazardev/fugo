package fg

import (
	"log"
	"strings"

	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// RouterWidget renders one route at a time and maintains a navigation history.
// Route patterns may contain :params (e.g. "/user/:id"); read captured values
// with Context.Param. Build one with Router.
type RouterWidget struct {
	routes        map[string]func() Widget
	current       string
	params        map[string]string
	history       []string
	currentWidget Widget
	beforeLeave   func() bool
	baseWidget
}

// Router creates a router from a map of route patterns to widget builders,
// starting on initialRoute. Patterns may contain :params, e.g. "/user/:id".
func Router(routes map[string]func() Widget, initialRoute string) *RouterWidget {
	return &RouterWidget{routes: routes, current: initialRoute}
}

// OnBeforeLeave registers a hook checked before every NavigateTo/GoBack; it
// returns false to cancel the navigation (e.g. to ask "discard unsaved
// changes?" before leaving a form) and returns the widget for chaining. A nil
// hook (the default) never blocks navigation.
func (r *RouterWidget) OnBeforeLeave(hook func() bool) *RouterWidget {
	r.beforeLeave = hook

	return r
}

// NavigateTo switches to path (matching a registered pattern, including
// :params), pushing the current route onto the history. It reports false if no
// pattern matches or OnBeforeLeave's hook vetoes the navigation.
func (r *RouterWidget) NavigateTo(path string) bool {
	_, params, ok := r.match(path)
	if !ok {
		log.Printf("[router] route not found: %s", path)

		return false
	}

	if r.beforeLeave != nil && !r.beforeLeave() {
		log.Printf("[router] navigate to %s vetoed by OnBeforeLeave", path)

		return false
	}

	if r.current != "" {
		r.history = append(r.history, r.current)
	}

	r.current = path
	r.params = params
	r.currentWidget = nil
	log.Printf("[router] navigate to: %s (history: %d)", path, len(r.history))

	return true
}

// GoBack pops the history and returns to the previous route. It reports false
// if the history is empty or OnBeforeLeave's hook vetoes the navigation.
func (r *RouterWidget) GoBack() bool {
	if len(r.history) == 0 {
		log.Println("[router] goback: no history")

		return false
	}

	if r.beforeLeave != nil && !r.beforeLeave() {
		log.Println("[router] goback vetoed by OnBeforeLeave")

		return false
	}

	r.current = r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
	_, r.params, _ = r.match(r.current)
	r.currentWidget = nil
	log.Printf("[router] goback to: %s (history: %d)", r.current, len(r.history))

	return true
}

// CurrentRoute returns the path currently being rendered.
func (r *RouterWidget) CurrentRoute() string { return r.current }

// Param returns the value captured for a :param in the current route (e.g. "id"
// for pattern "/user/:id"), or "" if there is no such parameter.
func (r *RouterWidget) Param(name string) string { return r.params[name] }

// match finds the builder whose pattern matches path, extracting any :params.
// An exact match takes precedence over a parameterized one.
func (r *RouterWidget) match(path string) (func() Widget, map[string]string, bool) {
	if b, ok := r.routes[path]; ok {
		return b, nil, true
	}

	pathSegs := splitSegs(path)
	for pattern, b := range r.routes {
		patSegs := splitSegs(pattern)
		if len(patSegs) != len(pathSegs) {
			continue
		}

		params := map[string]string{}
		matched := true

		for i, seg := range patSegs {
			switch {
			case strings.HasPrefix(seg, ":"):
				params[seg[1:]] = pathSegs[i]
			case seg != pathSegs[i]:
				matched = false
			}

			if !matched {
				break
			}
		}

		if matched {
			return b, params, true
		}
	}

	return nil, nil, false
}

func splitSegs(path string) []string {
	var segs []string

	for _, s := range strings.Split(path, "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}

	return segs
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

	const routerKey = "router"

	props, err := proto.Marshal(&fugov1.RouterProps{})
	if err != nil {
		flog.Errorf("marshal RouterProps: %v", err)
	}

	builder, params, ok := r.match(r.current)
	r.params = params

	if !ok {
		r.currentWidget = nil

		return []*fugov1.WidgetNode{{
			Id:    r.id,
			Key:   routerKey,
			Type:  fugov1.WidgetType_ROUTER,
			Props: props,
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
		Key:      routerKey,
		Type:     fugov1.WidgetType_ROUTER,
		Props:    props,
		Children: childIDs,
	}}, childNodes...)
}
