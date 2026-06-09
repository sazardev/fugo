package fg

// colorSet is the Material color palette. Access a color through the Colors
// var, e.g. fg.Colors.Amber — the Flutter equivalent of Colors.amber. Every
// field is a ready-to-use Color, so you rarely need a raw fg.Hex(...).
type colorSet struct {
	// Primary swatches (the Material "500" shade).
	Red        Color
	Pink       Color
	Purple     Color
	DeepPurple Color
	Indigo     Color
	Blue       Color
	LightBlue  Color
	Cyan       Color
	Teal       Color
	Green      Color
	LightGreen Color
	Lime       Color
	Yellow     Color
	Amber      Color
	Orange     Color
	DeepOrange Color
	Brown      Color
	Grey       Color
	BlueGrey   Color

	// Accent swatches (the vivid "A200" shade).
	RedAccent        Color
	PinkAccent       Color
	PurpleAccent     Color
	DeepPurpleAccent Color
	IndigoAccent     Color
	BlueAccent       Color
	LightBlueAccent  Color
	CyanAccent       Color
	TealAccent       Color
	GreenAccent      Color
	LightGreenAccent Color
	LimeAccent       Color
	YellowAccent     Color
	AmberAccent      Color
	OrangeAccent     Color
	DeepOrangeAccent Color

	// Greyscale ramp (Material grey 50–900) plus the absolutes.
	Grey50      Color
	Grey100     Color
	Grey200     Color
	Grey300     Color
	Grey400     Color
	Grey500     Color
	Grey600     Color
	Grey700     Color
	Grey800     Color
	Grey900     Color
	Black       Color
	White       Color
	Transparent Color
}

// Colors is the Material color palette, mirroring Flutter's Colors
// (fg.Colors.Amber). Values are the standard Material swatches.
var Colors = colorSet{
	Red:        Hex("#F44336"),
	Pink:       Hex("#E91E63"),
	Purple:     Hex("#9C27B0"),
	DeepPurple: Hex("#673AB7"),
	Indigo:     Hex("#3F51B5"),
	Blue:       Hex("#2196F3"),
	LightBlue:  Hex("#03A9F4"),
	Cyan:       Hex("#00BCD4"),
	Teal:       Hex("#009688"),
	Green:      Hex("#4CAF50"),
	LightGreen: Hex("#8BC34A"),
	Lime:       Hex("#CDDC39"),
	Yellow:     Hex("#FFEB3B"),
	Amber:      Hex("#FFC107"),
	Orange:     Hex("#FF9800"),
	DeepOrange: Hex("#FF5722"),
	Brown:      Hex("#795548"),
	Grey:       Hex("#9E9E9E"),
	BlueGrey:   Hex("#607D8B"),

	RedAccent:        Hex("#FF5252"),
	PinkAccent:       Hex("#FF4081"),
	PurpleAccent:     Hex("#E040FB"),
	DeepPurpleAccent: Hex("#7C4DFF"),
	IndigoAccent:     Hex("#536DFE"),
	BlueAccent:       Hex("#448AFF"),
	LightBlueAccent:  Hex("#40C4FF"),
	CyanAccent:       Hex("#18FFFF"),
	TealAccent:       Hex("#64FFDA"),
	GreenAccent:      Hex("#69F0AE"),
	LightGreenAccent: Hex("#B2FF59"),
	LimeAccent:       Hex("#EEFF41"),
	YellowAccent:     Hex("#FFFF00"),
	AmberAccent:      Hex("#FFD740"),
	OrangeAccent:     Hex("#FFAB40"),
	DeepOrangeAccent: Hex("#FF6E40"),

	Grey50:      Hex("#FAFAFA"),
	Grey100:     Hex("#F5F5F5"),
	Grey200:     Hex("#EEEEEE"),
	Grey300:     Hex("#E0E0E0"),
	Grey400:     Hex("#BDBDBD"),
	Grey500:     Hex("#9E9E9E"),
	Grey600:     Hex("#757575"),
	Grey700:     Hex("#616161"),
	Grey800:     Hex("#424242"),
	Grey900:     Hex("#212121"),
	Black:       Hex("#000000"),
	White:       Hex("#FFFFFF"),
	Transparent: Color{},
}
