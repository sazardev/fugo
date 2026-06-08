package style

type EdgeInsets struct {
	Top, Right, Bottom, Left float64
}

func EdgeAll(v float64) EdgeInsets {
	return EdgeInsets{Top: v, Right: v, Bottom: v, Left: v}
}

func EdgeSymmetric(horizontal, vertical float64) EdgeInsets {
	return EdgeInsets{
		Top:    vertical,
		Right:  horizontal,
		Bottom: vertical,
		Left:   horizontal,
	}
}

func EdgeOnly(top, right, bottom, left float64) EdgeInsets {
	return EdgeInsets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}
