package style

import (
	"fmt"
	"log"
)

const (
	maxRGB  = 255
	hexLen6 = 6
	hexLen8 = 8
)

// Color is an RGBA color with 8-bit channels. Build one with Hex, RGB, or RGBA.
type Color struct {
	R, G, B, A uint8
}

// RGBA returns a Color from explicit red, green, blue, and alpha channels.
func RGBA(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

// RGB returns a fully opaque Color from red, green, and blue channels
// (alpha is set to 255).
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: maxRGB}
}

// Hex parses a "#RRGGBB" or "#RRGGBBAA" string into a Color. The leading "#"
// is optional. A 6-digit value is treated as fully opaque. On an empty,
// malformed, or wrong-length input it logs the error and returns the zero
// Color.
func Hex(s string) Color {
	if len(s) == 0 {
		return Color{}
	}

	if s[0] == '#' {
		s = s[1:]
	}

	var r, g, b uint8
	var a uint8 = maxRGB

	switch len(s) {
	case hexLen6:
		if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); err != nil {
			log.Printf("style: invalid hex6 color: %s: %v", s, err)

			return Color{}
		}
	case hexLen8:
		if _, err := fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a); err != nil {
			log.Printf("style: invalid hex8 color: %s: %v", s, err)

			return Color{}
		}
	}

	return Color{R: r, G: g, B: b, A: a}
}

// WithAlpha returns a copy of the Color with its alpha channel set to a.
func (c Color) WithAlpha(a uint8) Color {
	c.A = a

	return c
}

// String returns the color as an uppercase hex string: "#RRGGBB" when the
// color is fully opaque, otherwise "#RRGGBBAA".
func (c Color) String() string {
	if c.A == maxRGB {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}

	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}
