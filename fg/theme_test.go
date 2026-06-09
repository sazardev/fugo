package fg

import (
	"testing"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

func TestThemeDefaultIsLight(t *testing.T) {
	defer UseTheme(LightTheme())

	UseTheme(LightTheme())
	if got, want := CurrentTheme().Colors.Background.String(), LightTheme().Colors.Background.String(); got != want {
		t.Errorf("default active theme background = %s, want light %s", got, want)
	}
}

// TestThemeLegacyDefaults guards the dark theme tokens against drift.
func TestThemeLegacyDefaults(t *testing.T) {
	d := DarkTheme()
	if d.Typography.Body != 14 {
		t.Errorf("Typography.Body = %v, want 14 (legacy default font size)", d.Typography.Body)
	}

	if d.Radius.MD != 8 {
		t.Errorf("Radius.MD = %v, want 8 (legacy default border radius)", d.Radius.MD)
	}

	if got, want := d.Colors.Primary.String(), Hex("#3B82F6").String(); got != want {
		t.Errorf("Colors.Primary = %s, want %s", got, want)
	}

	if got, want := d.Colors.OnSurface.String(), Hex("#FFFFFF").String(); got != want {
		t.Errorf("Colors.OnSurface = %s, want %s", got, want)
	}
}

func TestThemeBrightness(t *testing.T) {
	if got := LightTheme().Brightness(); got != brightnessLight {
		t.Errorf("LightTheme().Brightness() = %q, want light", got)
	}

	if got := DarkTheme().Brightness(); got != brightnessDark {
		t.Errorf("DarkTheme().Brightness() = %q, want dark", got)
	}
}

// Material 3 is native: a plain widget carries no color so the client's
// ColorScheme styles it; an explicit setter still wins.
func TestTextColorUnsetByDefault(t *testing.T) {
	if Text("hi").colorSet {
		t.Error("Text colorSet = true by default, want false (M3 styles it)")
	}

	c := Hex("#FF0000")
	if got, want := Text("hi").Color(c).Style.Color.String(), c.String(); got != want {
		t.Errorf("explicit Text color = %s, want %s", got, want)
	}
}

func TestButtonBgUnsetByDefault(t *testing.T) {
	b := Button("x")
	if b.bgColorSet {
		t.Error("Button bgColorSet = true by default, want false (M3 styles it)")
	}

	if b.variant != fugov1.ButtonVariant_BUTTON_FILLED {
		t.Errorf("Button variant = %v, want BUTTON_FILLED", b.variant)
	}

	c := Hex("#10B981")
	if got, want := Button("x").BgColor(c).bgColor.String(), c.String(); got != want {
		t.Errorf("explicit Button bg = %s, want %s", got, want)
	}
}

func TestButtonVariants(t *testing.T) {
	cases := []struct {
		w    *ButtonWidget
		want fugov1.ButtonVariant
	}{
		{FilledButton("a"), fugov1.ButtonVariant_BUTTON_FILLED},
		{FilledTonalButton("a"), fugov1.ButtonVariant_BUTTON_FILLED_TONAL},
		{OutlinedButton("a"), fugov1.ButtonVariant_BUTTON_OUTLINED},
		{TextButton("a"), fugov1.ButtonVariant_BUTTON_TEXT},
		{ElevatedButton("a"), fugov1.ButtonVariant_BUTTON_ELEVATED},
		{IconButton("home"), fugov1.ButtonVariant_BUTTON_ICON},
	}

	for _, tc := range cases {
		if tc.w.variant != tc.want {
			t.Errorf("variant = %v, want %v", tc.w.variant, tc.want)
		}
	}
}
