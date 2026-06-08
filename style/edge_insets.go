package style

// EdgeInsets describes spacing offsets for the four sides of a box, used for
// padding and margins. Build one with EdgeAll, EdgeSymmetric, or EdgeOnly.
type EdgeInsets struct {
	Top, Right, Bottom, Left float64
}

// EdgeAll returns insets with the same value v applied to all four sides.
func EdgeAll(v float64) EdgeInsets {
	return EdgeInsets{Top: v, Right: v, Bottom: v, Left: v}
}

// EdgeSymmetric returns insets with horizontal applied to the left and right
// sides and vertical applied to the top and bottom sides.
func EdgeSymmetric(horizontal, vertical float64) EdgeInsets {
	return EdgeInsets{
		Top:    vertical,
		Right:  horizontal,
		Bottom: vertical,
		Left:   horizontal,
	}
}

// EdgeOnly returns insets with each side set independently.
func EdgeOnly(top, right, bottom, left float64) EdgeInsets {
	return EdgeInsets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}
