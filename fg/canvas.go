package fg

import (
	"github.com/sazardev/fugo/flog"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// CanvasWidget is a small drawing surface for sparklines, gauges, and simple
// charts. It is a bounded, declarative subset of Canvas (line, rect, circle,
// path, filled path) — not general-purpose custom painting. Build one with
// Canvas and compose shapes onto it with Line/Rect/Circle/Path/FilledPath.
type CanvasWidget struct {
	width, height float64
	shapes        []*fugov1.ShapeSpec
	baseWidget
}

// Canvas creates a small drawing surface of the given size (logical pixels)
// for sparklines, gauges, and simple charts — a bounded, declarative subset
// of Canvas, not general-purpose custom painting. Compose shapes onto it with
// Line/Rect/Circle/Path/FilledPath, in the order they should be drawn (later
// ones paint over earlier ones).
func Canvas(width, height float64) *CanvasWidget {
	return &CanvasWidget{width: width, height: height}
}

// Line draws a straight line from (x1,y1) to (x2,y2) and returns the widget
// for chaining. color is a hex string (e.g. "#3B82F6"); strokeWidth <= 0 uses
// a 1px default.
func (w *CanvasWidget) Line(x1, y1, x2, y2 float64, color string, strokeWidth float64) *CanvasWidget {
	w.shapes = append(w.shapes, &fugov1.ShapeSpec{
		Kind: fugov1.ShapeKind_SHAPE_LINE,
		Points: []*fugov1.Point{
			{X: x1, Y: y1},
			{X: x2, Y: y2},
		},
		Color:       color,
		StrokeWidth: strokeWidth,
	})

	return w
}

// Rect draws a rectangle from top-left (x1,y1) to bottom-right (x2,y2).
// fillColor "" means no fill (stroke only).
func (w *CanvasWidget) Rect(x1, y1, x2, y2 float64, color, fillColor string, strokeWidth float64) *CanvasWidget {
	w.shapes = append(w.shapes, &fugov1.ShapeSpec{
		Kind: fugov1.ShapeKind_SHAPE_RECT,
		Points: []*fugov1.Point{
			{X: x1, Y: y1},
			{X: x2, Y: y2},
		},
		Color:       color,
		FillColor:   fillColor,
		StrokeWidth: strokeWidth,
	})

	return w
}

// Circle draws a circle centered at (cx,cy) with the given radius.
func (w *CanvasWidget) Circle(cx, cy, radius float64, color, fillColor string, strokeWidth float64) *CanvasWidget {
	w.shapes = append(w.shapes, &fugov1.ShapeSpec{
		Kind:        fugov1.ShapeKind_SHAPE_CIRCLE,
		Points:      []*fugov1.Point{{X: cx, Y: cy}},
		Radius:      radius,
		Color:       color,
		FillColor:   fillColor,
		StrokeWidth: strokeWidth,
	})

	return w
}

// Path draws an open polyline through points, in order. Each point is an
// [x, y] pair, e.g. fg.Canvas(100, 40).Path([][2]float64{{0, 20}, {50, 5}, {100, 30}}, "#3B82F6", 2).
func (w *CanvasWidget) Path(points [][2]float64, color string, strokeWidth float64) *CanvasWidget {
	w.shapes = append(w.shapes, &fugov1.ShapeSpec{
		Kind:        fugov1.ShapeKind_SHAPE_PATH,
		Points:      toPoints(points),
		Color:       color,
		StrokeWidth: strokeWidth,
	})

	return w
}

// FilledPath draws a closed, filled polygon through points, in order. Each
// point is an [x, y] pair (see Path for the format).
func (w *CanvasWidget) FilledPath(points [][2]float64, color, fillColor string) *CanvasWidget {
	w.shapes = append(w.shapes, &fugov1.ShapeSpec{
		Kind:      fugov1.ShapeKind_SHAPE_FILLED_PATH,
		Points:    toPoints(points),
		Color:     color,
		FillColor: fillColor,
	})

	return w
}

func toPoints(points [][2]float64) []*fugov1.Point {
	out := make([]*fugov1.Point, 0, len(points))
	for _, p := range points {
		out = append(out, &fugov1.Point{X: p[0], Y: p[1]})
	}

	return out
}

func (w *CanvasWidget) isWidget()                {}
func (w *CanvasWidget) widgetChildren() []Widget { return nil }

func (w *CanvasWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	w.id = *counter

	props, err := proto.Marshal(&fugov1.CanvasProps{
		Width:  w.width,
		Height: w.height,
		Shapes: w.shapes,
	})
	if err != nil {
		flog.Errorf("marshal CanvasProps: %v", err)
	}

	return []*fugov1.WidgetNode{{
		Id:    w.id,
		Key:   w.key,
		Type:  WidgetCanvas,
		Props: props,
	}}
}
