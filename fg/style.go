package fg

import "github.com/sazardev/fugo/style"

// Convenience aliases for the most-used types from the style package, so app
// code can stay in the fg namespace (fg.Color, fg.EdgeInsets, ...).
type (
	// Color is an RGBA color (alias of style.Color).
	Color = style.Color
	// EdgeInsets are padding/margin offsets (alias of style.EdgeInsets).
	EdgeInsets = style.EdgeInsets
	// FontWeight is a text weight (alias of style.FontWeight).
	FontWeight = style.FontWeight
	// TextAlign is a horizontal text alignment (alias of style.TextAlign).
	TextAlign = style.TextAlign
	// TextStyle describes text rendering (alias of style.TextStyle).
	TextStyle = style.TextStyle
	// Border describes a box border (alias of style.Border).
	Border = style.Border
	// BorderSide describes one side of a border (alias of style.BorderSide).
	BorderSide = style.BorderSide
)

// Re-exported style values for common font weights and text alignments.
var (
	WeightNormal = style.WeightNormal
	WeightBold   = style.WeightBold
	AlignLeft    = style.AlignLeft
	AlignCenter  = style.AlignCenter
	AlignRight   = style.AlignRight
)

// Hex parses a "#RRGGBB" or "#RRGGBBAA" color string. See [style.Hex].
func Hex(s string) Color { return style.Hex(s) }

// RGB builds an opaque color from 8-bit channels. See [style.RGB].
func RGB(r, g, b uint8) Color { return style.RGB(r, g, b) }

// RGBA builds a color from 8-bit channels including alpha. See [style.RGBA].
func RGBA(r, g, b, a uint8) Color { return style.RGBA(r, g, b, a) }

// EdgeAll returns insets with the same value on all four sides.
func EdgeAll(v float64) EdgeInsets { return style.EdgeAll(v) }

// EdgeSymmetric returns insets with horizontal (h) and vertical (v) values.
func EdgeSymmetric(h, v float64) EdgeInsets { return style.EdgeSymmetric(h, v) }

// EdgeOnly returns insets with independent top, right, bottom, and left values.
func EdgeOnly(top, right, bottom, left float64) EdgeInsets {
	return style.EdgeOnly(top, right, bottom, left)
}

// NewTextStyle builds a TextStyle with the given font size and color.
func NewTextStyle(fontSize float64, color Color) TextStyle {
	return style.NewTextStyle(fontSize, color)
}

// BorderAll returns a uniform border of the given color and width.
func BorderAll(c Color, width float64) Border { return style.BorderAll(c, width) }
