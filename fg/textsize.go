package fg

// textScale is the Material 3 type scale in logical pixels. Access a size
// through the TextSize var, e.g. fg.Text("Hi").FontSize(fg.TextSize.HeadlineMedium),
// so layouts use consistent, named sizes instead of magic numbers.
type textScale struct {
	DisplayLarge  float64
	DisplayMedium float64
	DisplaySmall  float64

	HeadlineLarge  float64
	HeadlineMedium float64
	HeadlineSmall  float64

	TitleLarge  float64
	TitleMedium float64
	TitleSmall  float64

	BodyLarge  float64
	BodyMedium float64
	BodySmall  float64

	LabelLarge  float64
	LabelMedium float64
	LabelSmall  float64
}

// TextSize is the Material 3 type scale (fg.TextSize.HeadlineMedium).
var TextSize = textScale{
	DisplayLarge:  57,
	DisplayMedium: 45,
	DisplaySmall:  36,

	HeadlineLarge:  32,
	HeadlineMedium: 28,
	HeadlineSmall:  24,

	TitleLarge:  22,
	TitleMedium: 16,
	TitleSmall:  14,

	BodyLarge:  16,
	BodyMedium: 14,
	BodySmall:  12,

	LabelLarge:  14,
	LabelMedium: 12,
	LabelSmall:  11,
}
