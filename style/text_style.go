package style

type FontWeight int

const (
	WeightNormal FontWeight = 400
	WeightBold   FontWeight = 700
)

type TextAlign int

const (
	AlignLeft   TextAlign = 0
	AlignCenter TextAlign = 1
	AlignRight  TextAlign = 2
)

type TextStyle struct {
	Color    Color
	FontSize float64
	Weight   FontWeight
	Align    TextAlign
}

func NewTextStyle(fontSize float64, color Color) TextStyle {
	return TextStyle{
		Color:    color,
		FontSize: fontSize,
		Weight:   WeightNormal,
	}
}

func (t TextStyle) WithWeight(w FontWeight) TextStyle {
	t.Weight = w

	return t
}

func (t TextStyle) WithAlign(a TextAlign) TextStyle {
	t.Align = a

	return t
}
