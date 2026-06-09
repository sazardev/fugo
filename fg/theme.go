package fg

// Theme holds the opinionated design tokens that widget constructors read from
// when no explicit override is set. Set the active theme once at startup with
// UseTheme; widgets created afterwards inherit its tokens. Any chainable setter
// (e.g. .Color(), .BgColor()) still overrides the theme on a per-widget basis.
type Theme struct {
	Colors     ThemeColors
	Typography ThemeTypography
	Spacing    ThemeSpacing
	Radius     ThemeRadius
	// Dark selects the Material 3 brightness the Flutter client builds its
	// ColorScheme with (seeded from Colors.Primary). LightTheme sets it false,
	// DarkTheme true.
	Dark bool
}

// Material brightness values, as the Flutter client expects them over
// FUGO_THEME_BRIGHTNESS.
const (
	brightnessDark  = "dark"
	brightnessLight = "light"
)

// Brightness reports the Material brightness for the active theme.
func (t Theme) Brightness() string {
	if t.Dark {
		return brightnessDark
	}

	return brightnessLight
}

// ThemeColors are the semantic color roles of a theme.
type ThemeColors struct {
	Primary    Color // interactive elements: buttons, accents
	Secondary  Color // secondary actions
	Background Color // app background
	Surface    Color // cards, panels, sheets
	Error      Color // destructive / error states
	Success    Color // positive / success states
	OnPrimary  Color // content rendered on top of Primary
	OnSurface  Color // primary text/icons on Background/Surface
	Muted      Color // secondary / disabled text
	Border     Color // dividers and outlines
}

// ThemeTypography holds the default text sizes (logical px) and weight.
type ThemeTypography struct {
	Family  string
	Heading float64
	Body    float64
	Caption float64
	Weight  FontWeight
}

// ThemeSpacing is a 5-step spacing scale in logical pixels.
type ThemeSpacing struct {
	XS, SM, MD, LG, XL float64
}

// ThemeRadius is a 3-step corner-radius scale in logical pixels.
type ThemeRadius struct {
	SM, MD, LG float64
}

// DarkTheme is the opinionated dark theme. Activate it with UseTheme before
// calling RunStandalone (the client reads the brightness/seed at startup).
func DarkTheme() Theme {
	return Theme{
		Colors: ThemeColors{
			Primary:    Hex("#3B82F6"),
			Secondary:  Hex("#8B5CF6"),
			Background: Hex("#1A1A2E"),
			Surface:    Hex("#16213E"),
			Error:      Hex("#EF4444"),
			Success:    Hex("#10B981"),
			OnPrimary:  Hex("#FFFFFF"),
			OnSurface:  Hex("#FFFFFF"),
			Muted:      Hex("#9CA3AF"),
			Border:     Hex("#6B7280"),
		},
		Typography: defaultTypography(),
		Spacing:    defaultSpacing(),
		Radius:     defaultRadius(),
		Dark:       true,
	}
}

// LightTheme is the opinionated default light theme (active out of the box).
func LightTheme() Theme {
	return Theme{
		Colors: ThemeColors{
			Primary:    Hex("#2563EB"),
			Secondary:  Hex("#7C3AED"),
			Background: Hex("#FFFFFF"),
			Surface:    Hex("#F3F4F6"),
			Error:      Hex("#DC2626"),
			Success:    Hex("#059669"),
			OnPrimary:  Hex("#FFFFFF"),
			OnSurface:  Hex("#111827"),
			Muted:      Hex("#6B7280"),
			Border:     Hex("#D1D5DB"),
		},
		Typography: defaultTypography(),
		Spacing:    defaultSpacing(),
		Radius:     defaultRadius(),
	}
}

func defaultTypography() ThemeTypography {
	return ThemeTypography{
		Family:  "",
		Heading: 20,
		Body:    14,
		Caption: 12,
		Weight:  WeightNormal,
	}
}

func defaultSpacing() ThemeSpacing {
	return ThemeSpacing{XS: 4, SM: 8, MD: 16, LG: 24, XL: 32}
}

func defaultRadius() ThemeRadius {
	return ThemeRadius{SM: 4, MD: 8, LG: 16}
}

//nolint:gochecknoglobals // the active theme is intentional, opinionated, package-level state
var active = LightTheme()

// UseTheme sets the active theme. Call once at startup, before building the UI;
// widget constructors created afterwards inherit its tokens.
func UseTheme(t Theme) { active = t }

// CurrentTheme returns the active theme, handy for reading tokens while
// composing UI (e.g. fg.CurrentTheme().Spacing.MD).
func CurrentTheme() Theme { return active }
