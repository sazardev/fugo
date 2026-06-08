package style

type BorderSide struct {
	Color Color
	Width float64
}

type Border struct {
	Top, Right, Bottom, Left BorderSide
}

func BorderAll(color Color, width float64) Border {
	s := BorderSide{Color: color, Width: width}

	return Border{Top: s, Right: s, Bottom: s, Left: s}
}
