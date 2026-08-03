package fg

import (
	"testing"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// TestCanvasWalkNodes ensures Canvas produces a single CANVAS node whose
// props carry the width/height and every shape added via the builder methods.
func TestCanvasWalkNodes(t *testing.T) {
	c := Canvas(100, 40).
		Line(0, 0, 10, 10, "#111111", 2).
		Rect(0, 0, 20, 20, "#222222", "#333333", 1).
		Circle(50, 20, 5, "#444444", "", 0).
		Path([][2]float64{{0, 20}, {50, 5}, {100, 30}}, "#555555", 1.5).
		FilledPath([][2]float64{{0, 0}, {10, 0}, {10, 10}}, "#666666", "#777777")

	tree, _ := BuildTree(c)

	node := nodeOfType(tree, fugov1.WidgetType_CANVAS)
	if node == nil {
		t.Fatal("no CANVAS node produced")
	}

	var props fugov1.CanvasProps
	if err := proto.Unmarshal(node.GetProps(), &props); err != nil {
		t.Fatalf("unmarshal CanvasProps: %v", err)
	}

	if props.GetWidth() != 100 || props.GetHeight() != 40 {
		t.Errorf("size = (%v,%v), want (100,40)", props.GetWidth(), props.GetHeight())
	}

	shapes := props.GetShapes()
	if len(shapes) != 5 {
		t.Fatalf("got %d shapes, want 5", len(shapes))
	}

	line := shapes[0]
	if line.GetKind() != fugov1.ShapeKind_SHAPE_LINE {
		t.Errorf("shape0 kind = %v, want SHAPE_LINE", line.GetKind())
	}

	if line.GetColor() != "#111111" || line.GetStrokeWidth() != 2 {
		t.Errorf("line color/stroke = %q/%v, want #111111/2", line.GetColor(), line.GetStrokeWidth())
	}

	if len(line.GetPoints()) != 2 || line.GetPoints()[1].GetX() != 10 || line.GetPoints()[1].GetY() != 10 {
		t.Errorf("line points = %v, want [(0,0) (10,10)]", line.GetPoints())
	}

	rect := shapes[1]
	if rect.GetKind() != fugov1.ShapeKind_SHAPE_RECT {
		t.Errorf("shape1 kind = %v, want SHAPE_RECT", rect.GetKind())
	}

	if rect.GetFillColor() != "#333333" {
		t.Errorf("rect fill = %q, want #333333", rect.GetFillColor())
	}

	circle := shapes[2]
	if circle.GetKind() != fugov1.ShapeKind_SHAPE_CIRCLE {
		t.Errorf("shape2 kind = %v, want SHAPE_CIRCLE", circle.GetKind())
	}

	if circle.GetRadius() != 5 {
		t.Errorf("circle radius = %v, want 5", circle.GetRadius())
	}

	if len(circle.GetPoints()) != 1 || circle.GetPoints()[0].GetX() != 50 || circle.GetPoints()[0].GetY() != 20 {
		t.Errorf("circle center = %v, want (50,20)", circle.GetPoints())
	}

	path := shapes[3]
	if path.GetKind() != fugov1.ShapeKind_SHAPE_PATH {
		t.Errorf("shape3 kind = %v, want SHAPE_PATH", path.GetKind())
	}

	if len(path.GetPoints()) != 3 {
		t.Errorf("path points = %d, want 3", len(path.GetPoints()))
	}

	filled := shapes[4]
	if filled.GetKind() != fugov1.ShapeKind_SHAPE_FILLED_PATH {
		t.Errorf("shape4 kind = %v, want SHAPE_FILLED_PATH", filled.GetKind())
	}

	if filled.GetFillColor() != "#777777" {
		t.Errorf("filled path fill = %q, want #777777", filled.GetFillColor())
	}

	if node.GetId() == 0 {
		t.Error("node id should be assigned (non-zero)")
	}
}
