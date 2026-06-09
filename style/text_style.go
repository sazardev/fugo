package style

// FontWeight is the thickness of a font's glyphs, expressed on the standard
// 100-900 numeric scale (e.g. 400 for normal, 700 for bold).
type FontWeight int

// Common font weights.
const (
	WeightNormal FontWeight = 400 // WeightNormal is the regular (non-bold) weight.
	WeightBold   FontWeight = 700 // WeightBold is the bold weight.
)

// TextAlign controls the horizontal alignment of text within its container.
type TextAlign int

// Text alignment options.
const (
	AlignLeft   TextAlign = 0 // AlignLeft aligns text to the left edge.
	AlignCenter TextAlign = 1 // AlignCenter centers text horizontally.
	AlignRight  TextAlign = 2 // AlignRight aligns text to the right edge.
)

// TextStyle describes how text is rendered: its color, size, weight, and
// alignment. Build one with NewTextStyle and refine it with the With* methods.
type TextStyle struct {
	Color    Color
	FontSize float64
	Weight   FontWeight
	Align    TextAlign
}

// NewTextStyle returns a TextStyle with the given font size and color,
// defaulting to WeightNormal weight and left (zero) alignment.
func NewTextStyle(fontSize float64, color Color) TextStyle {
	return TextStyle{
		Color:    color,
		FontSize: fontSize,
		Weight:   WeightNormal,
	}
}

// WithWeight returns a copy of the TextStyle with its font weight set to w.
func (t TextStyle) WithWeight(w FontWeight) TextStyle {
	t.Weight = w

	return t
}

// WithAlign returns a copy of the TextStyle with its alignment set to a.
func (t TextStyle) WithAlign(a TextAlign) TextStyle {
	t.Align = a

	return t
}
