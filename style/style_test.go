package style

import (
	"testing"
)

func TestColor_Hex(t *testing.T) {
	c := Hex("#FF0000")
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Errorf("Hex(#FF0000) = %+v, want R=255", c)
	}

	c = Hex("#00FF00")
	if c.R != 0 || c.G != 255 || c.B != 0 {
		t.Errorf("Hex(#00FF00) = %+v, want G=255", c)
	}

	c = Hex("#0000FF")
	if c.B != 255 {
		t.Errorf("Hex(#0000FF) = %+v, want B=255", c)
	}
}

func TestColor_HexWithAlpha(t *testing.T) {
	c := Hex("#FF000080")
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 128 {
		t.Errorf("Hex(#FF000080) = %+v, want A=128", c)
	}
}

func TestColor_HexInvalid(t *testing.T) {
	c := Hex("invalid")
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("invalid hex should return zero Color for RGB")
	}
}

func TestColor_RGB_RGBA(t *testing.T) {
	c := RGB(10, 20, 30)
	if c.R != 10 || c.G != 20 || c.B != 30 || c.A != 255 {
		t.Errorf("RGB(10,20,30) = %+v", c)
	}

	c = RGBA(10, 20, 30, 100)
	if c.R != 10 || c.G != 20 || c.B != 30 || c.A != 100 {
		t.Errorf("RGBA(10,20,30,100) = %+v", c)
	}
}

func TestColor_String(t *testing.T) {
	c := RGB(255, 0, 0)
	if s := c.String(); s != "#FF0000" {
		t.Errorf("String() = %s, want #FF0000", s)
	}

	c = RGBA(255, 0, 0, 128)
	if s := c.String(); s != "#FF000080" {
		t.Errorf("String() = %s, want #FF000080", s)
	}
}

func TestColor_WithAlpha(t *testing.T) {
	c := RGB(255, 0, 0).WithAlpha(128)
	if c.A != 128 {
		t.Errorf("WithAlpha(128) A = %d", c.A)
	}
}

func TestEdgeInsets_Constructors(t *testing.T) {
	e := EdgeAll(10)
	if e.Top != 10 || e.Right != 10 || e.Bottom != 10 || e.Left != 10 {
		t.Errorf("EdgeAll(10) = %+v", e)
	}

	e = EdgeSymmetric(10, 20)
	if e.Top != 20 || e.Right != 10 || e.Bottom != 20 || e.Left != 10 {
		t.Errorf("EdgeSymmetric(10,20) = %+v", e)
	}

	e = EdgeOnly(1, 2, 3, 4)
	if e.Top != 1 || e.Right != 2 || e.Bottom != 3 || e.Left != 4 {
		t.Errorf("EdgeOnly(1,2,3,4) = %+v", e)
	}
}

func TestTextStyle(t *testing.T) {
	ts := NewTextStyle(14, Hex("#FFFFFF"))
	if ts.FontSize != 14 {
		t.Errorf("FontSize = %f, want 14", ts.FontSize)
	}
	if ts.Weight != WeightNormal {
		t.Errorf("Weight = %d, want Normal", ts.Weight)
	}

	ts = ts.WithWeight(WeightBold)
	if ts.Weight != WeightBold {
		t.Errorf("WithWeight = %d, want Bold", ts.Weight)
	}

	ts = ts.WithAlign(AlignCenter)
	if ts.Align != AlignCenter {
		t.Errorf("WithAlign = %d, want Center", ts.Align)
	}
}
