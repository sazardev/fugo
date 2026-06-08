package fg

import (
	"testing"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

func newTestRouter() *RouterWidget {
	return Router(map[string]func() Widget{
		"/":     func() Widget { return Text("home") },
		"/next": func() Widget { return Text("next") },
	}, "/")
}

func TestRouterInitialRoute(t *testing.T) {
	r := newTestRouter()
	if r.CurrentRoute() != "/" {
		t.Errorf("initial route = %s, want /", r.CurrentRoute())
	}
}

func TestRouterNavigateTo(t *testing.T) {
	r := newTestRouter()
	if !r.NavigateTo("/next") {
		t.Fatal("NavigateTo(/next) should succeed")
	}
	if r.CurrentRoute() != "/next" {
		t.Errorf("route = %s, want /next", r.CurrentRoute())
	}
}

func TestRouterNavigateUnknown(t *testing.T) {
	r := newTestRouter()
	if r.NavigateTo("/missing") {
		t.Error("NavigateTo to an unregistered route should return false")
	}
	if r.CurrentRoute() != "/" {
		t.Errorf("route should stay /, got %s", r.CurrentRoute())
	}
}

func TestRouterGoBack(t *testing.T) {
	r := newTestRouter()
	r.NavigateTo("/next")

	if !r.GoBack() {
		t.Error("GoBack should succeed when history is non-empty")
	}
	if r.CurrentRoute() != "/" {
		t.Errorf("after GoBack route = %s, want /", r.CurrentRoute())
	}
	if r.GoBack() {
		t.Error("GoBack with empty history should return false")
	}
}

func TestRouterUnmatchedRouteRendersPlaceholder(t *testing.T) {
	r := Router(map[string]func() Widget{
		"/": func() Widget { return Text("home") },
	}, "/missing")

	var counter uint32
	nodes := r.walkNodes(&counter)

	if len(nodes) != 1 || nodes[0].GetType() != fugov1.WidgetType_CONTAINER {
		t.Errorf("an unmatched route should render a single placeholder container, got %d nodes", len(nodes))
	}
}

func TestRouterParams(t *testing.T) {
	r := Router(map[string]func() Widget{
		"/":         func() Widget { return Text("home") },
		"/user/:id": func() Widget { return Text("user") },
	}, "/")

	if !r.NavigateTo("/user/42") {
		t.Fatal("NavigateTo(/user/42) should match /user/:id")
	}
	if got := r.Param("id"); got != "42" {
		t.Errorf("Param(id) = %q, want 42", got)
	}
	if r.CurrentRoute() != "/user/42" {
		t.Errorf("CurrentRoute = %q, want /user/42", r.CurrentRoute())
	}
}

func TestRouterWalkNodesRendersCurrent(t *testing.T) {
	r := newTestRouter()

	var counter uint32
	nodes := r.walkNodes(&counter)

	// router wrapper node + the current page's Text node
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes (router + current page), got %d", len(nodes))
	}
}
