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

type Color struct {
	R, G, B, A uint8
}

func RGBA(r, g, b, a uint8) Color {
	return Color{R: r, G: g, B: b, A: a}
}

func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, A: maxRGB}
}

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

func (c Color) WithAlpha(a uint8) Color {
	c.A = a

	return c
}

func (c Color) String() string {
	if c.A == maxRGB {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}

	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}
