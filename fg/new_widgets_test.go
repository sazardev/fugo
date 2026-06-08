package fg

import (
	"testing"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

func TestSpacerWrapsExpanded(t *testing.T) {
	var counter uint32
	nodes := Spacer().walkNodes(&counter)

	if len(nodes) != 2 { // expanded + empty sizedbox
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].GetType() != fugov1.WidgetType_EXPANDED {
		t.Errorf("root = %v, want EXPANDED", nodes[0].GetType())
	}
}

func TestScrollViewWalkNodes(t *testing.T) {
	var counter uint32
	nodes := ScrollView(Text("x")).Horizontal().walkNodes(&counter)

	if nodes[0].GetType() != fugov1.WidgetType_SCROLLVIEW {
		t.Errorf("type = %v, want SCROLLVIEW", nodes[0].GetType())
	}
	if len(nodes[0].GetChildren()) != 1 {
		t.Errorf("expected 1 child, got %d", len(nodes[0].GetChildren()))
	}
}

func TestAlignWalkNodes(t *testing.T) {
	var counter uint32
	nodes := Align(Text("x"), 0.5, -0.5).walkNodes(&counter)

	if nodes[0].GetType() != fugov1.WidgetType_ALIGN {
		t.Errorf("type = %v, want ALIGN", nodes[0].GetType())
	}
}

func TestGestureDetectorHandler(t *testing.T) {
	tapped := false
	g := GestureDetector(Text("x")).OnTap(func(Event) { tapped = true })

	if !g.HasHandler() {
		t.Fatal("expected a handler")
	}

	g.Handle(Event{})
	if !tapped {
		t.Error("handler was not invoked")
	}

	var counter uint32
	if nodes := g.walkNodes(&counter); nodes[0].GetType() != fugov1.WidgetType_GESTUREDETECTOR {
		t.Errorf("type = %v, want GESTUREDETECTOR", nodes[0].GetType())
	}
}

func TestRadioWalkNodes(t *testing.T) {
	r := Radio("a", "Option A").Group("sel").OnChange(func(Event) {})

	if !r.HasHandler() {
		t.Error("expected a handler")
	}

	var counter uint32
	if nodes := r.walkNodes(&counter); nodes[0].GetType() != fugov1.WidgetType_RADIO {
		t.Errorf("type = %v, want RADIO", nodes[0].GetType())
	}
}

func TestDropdownWalkNodes(t *testing.T) {
	d := Dropdown([]string{"a", "b"}).SetValue("a").OnChange(func(Event) {})

	if !d.HasHandler() {
		t.Error("expected a handler")
	}

	var counter uint32
	if nodes := d.walkNodes(&counter); nodes[0].GetType() != fugov1.WidgetType_DROPDOWN {
		t.Errorf("type = %v, want DROPDOWN", nodes[0].GetType())
	}
}
