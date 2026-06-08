package fg

import "testing"

func TestThemeDefaultIsDark(t *testing.T) {
	got := CurrentTheme().Colors.Background.String()
	want := DarkTheme().Colors.Background.String()
	if got != want {
		t.Errorf("default active theme background = %s, want dark %s", got, want)
	}
}

// TestThemeLegacyDefaults guards against visual regressions: the dark theme
// tokens must equal the values that widget constructors hardcoded before
// theming existed.
func TestThemeLegacyDefaults(t *testing.T) {
	d := DarkTheme()
	if d.Typography.Body != 14 {
		t.Errorf("Typography.Body = %v, want 14 (legacy default font size)", d.Typography.Body)
	}
	if d.Radius.MD != 8 {
		t.Errorf("Radius.MD = %v, want 8 (legacy default border radius)", d.Radius.MD)
	}
	if got, want := d.Colors.Primary.String(), Hex("#3B82F6").String(); got != want {
		t.Errorf("Colors.Primary = %s, want %s (legacy button bg)", got, want)
	}
	if got, want := d.Colors.OnSurface.String(), Hex("#FFFFFF").String(); got != want {
		t.Errorf("Colors.OnSurface = %s, want %s (legacy text color)", got, want)
	}
}

func TestThemeTextUsesActiveTheme(t *testing.T) {
	defer UseTheme(DarkTheme())

	UseTheme(LightTheme())
	if got, want := Text("hi").Style.Color.String(), LightTheme().Colors.OnSurface.String(); got != want {
		t.Errorf("Text color under light theme = %s, want %s", got, want)
	}
}

func TestThemeButtonUsesActiveTheme(t *testing.T) {
	defer UseTheme(DarkTheme())

	UseTheme(LightTheme())
	if got, want := Button("x").bgColor.String(), LightTheme().Colors.Primary.String(); got != want {
		t.Errorf("Button bg under light theme = %s, want %s", got, want)
	}

	UseTheme(DarkTheme())
	if got, want := Button("x").bgColor.String(), DarkTheme().Colors.Primary.String(); got != want {
		t.Errorf("Button bg under dark theme = %s, want %s", got, want)
	}
}
