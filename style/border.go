package style

// BorderSide describes the color and width of a single edge of a border.
type BorderSide struct {
	Color Color
	Width float64
}

// Border describes the four edges of a box border, each as its own BorderSide.
// Build a uniform border with BorderAll.
type Border struct {
	Top, Right, Bottom, Left BorderSide
}

// BorderAll returns a Border with the same color and width applied to all four
// sides.
func BorderAll(color Color, width float64) Border {
	s := BorderSide{Color: color, Width: width}

	return Border{Top: s, Right: s, Bottom: s, Left: s}
}
