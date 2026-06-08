package fg

import (
	"testing"

	"github.com/sazardev/fugo/style"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

func TestBuildTree(t *testing.T) {
	root := Text("hello")
	tree, m := BuildTree(root)

	if tree == nil {
		t.Fatal("BuildTree returned nil")
	}
	if tree.GetRoot() == 0 {
		t.Error("Root should be non-zero")
	}
	if len(tree.GetNodes()) != 1 {
		t.Errorf("Expected 1 node, got %d", len(tree.GetNodes()))
	}
	if len(m) != 1 {
		t.Errorf("Expected 1 widget in map, got %d", len(m))
	}
}

func TestBuildTreeNested(t *testing.T) {
	root := Container(
		Column(
			Text("a"),
			Text("b"),
		),
	)

	tree, _ := BuildTree(root)
	if len(tree.GetNodes()) != 4 {
		t.Errorf("Expected 4 nodes (container+column+2text), got %d", len(tree.GetNodes()))
	}
	if tree.GetRoot() != tree.GetNodes()[0].GetId() {
		t.Error("Root should point to first node")
	}
}

func TestWithKey(t *testing.T) {
	w := Text("hello")
	if w.widgetKey() != "" {
		t.Error("Key should be empty initially")
	}

	WithKey(w, "mykey")
	if w.widgetKey() != "mykey" {
		t.Errorf("Key = %s, want mykey", w.widgetKey())
	}
}

func TestTextWalkNodes(t *testing.T) {
	txt := Text("hello")
	txt.Style = style.NewTextStyle(16, style.Hex("#FFFFFF"))

	var counter uint32
	nodes := txt.walkNodes(&counter)

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodes))
	}

	n := nodes[0]
	if n.GetType() != fugov1.WidgetType_TEXT {
		t.Errorf("Type = %v, want TEXT", n.GetType())
	}
	if n.GetKey() != "" {
		t.Errorf("Key should be empty, got %s", n.GetKey())
	}
}

func TestButtonWalkNodes(t *testing.T) {
	btn := Button("Click")
	btn.OnClick(func(_ Event) {})

	var counter uint32
	nodes := btn.walkNodes(&counter)

	n := nodes[0]
	if n.GetType() != fugov1.WidgetType_BUTTON {
		t.Errorf("Type = %v, want BUTTON", n.GetType())
	}
}

func TestColumnWalkNodes(t *testing.T) {
	col := Column(
		Text("a"),
		Text("b"),
	)

	var counter uint32
	nodes := col.walkNodes(&counter)

	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
	root := nodes[0]
	if root.GetType() != fugov1.WidgetType_COLUMN {
		t.Errorf("Type = %v, want COLUMN", root.GetType())
	}
	if len(root.GetChildren()) != 2 {
		t.Errorf("Expected 2 children, got %d", len(root.GetChildren()))
	}
}

func TestRowWalkNodes(t *testing.T) {
	row := Row(Text("a"), Text("b"))

	var counter uint32
	nodes := row.walkNodes(&counter)

	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
}

func TestStackWalkNodes(t *testing.T) {
	stack := Stack(Text("a"), Text("b"))

	var counter uint32
	nodes := stack.walkNodes(&counter)

	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
}

func TestContainerWalkNodes(t *testing.T) {
	cont := Container(Text("hello"))
	cont.BgColor(style.Hex("#FF0000"))

	var counter uint32
	nodes := cont.walkNodes(&counter)

	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}
}
