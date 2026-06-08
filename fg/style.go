package fg

import "github.com/sazardev/fugo/style"

type (
	Color      = style.Color
	EdgeInsets = style.EdgeInsets
	FontWeight = style.FontWeight
	TextAlign  = style.TextAlign
	TextStyle  = style.TextStyle
	Border     = style.Border
	BorderSide = style.BorderSide
)

var (
	WeightNormal = style.WeightNormal
	WeightBold   = style.WeightBold
	AlignLeft    = style.AlignLeft
	AlignCenter  = style.AlignCenter
	AlignRight   = style.AlignRight
)

func Hex(s string) Color                    { return style.Hex(s) }
func RGB(r, g, b uint8) Color               { return style.RGB(r, g, b) }
func RGBA(r, g, b, a uint8) Color           { return style.RGBA(r, g, b, a) }
func EdgeAll(v float64) EdgeInsets          { return style.EdgeAll(v) }
func EdgeSymmetric(h, v float64) EdgeInsets { return style.EdgeSymmetric(h, v) }
func EdgeOnly(top, right, bottom, left float64) EdgeInsets {
	return style.EdgeOnly(top, right, bottom, left)
}

func NewTextStyle(fontSize float64, color Color) TextStyle {
	return style.NewTextStyle(fontSize, color)
}
func BorderAll(c Color, width float64) Border { return style.BorderAll(c, width) }
