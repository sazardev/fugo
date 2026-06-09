package fg

import "testing"

func TestIconsTable(t *testing.T) {
	cases := []struct {
		got, want string
	}{
		{Icons.Home, "home"},
		{Icons.Add, "add"},
		{Icons.Settings, "settings"},
		{Icons.ArrowBack, "arrow_back"},
		{Icons.Delete, "delete"},
		// Icons beyond the small set the client used to hardcode — proof the
		// full generated Material table is wired up.
		{Icons.Coffee, "coffee"},
		{Icons.Email, "email"},
		{Icons.Bolt, "bolt"},
	}

	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("icon = %q, want %q", c.got, c.want)
		}
	}
}

func TestColorsPalette(t *testing.T) {
	if got := Colors.Amber.String(); got != "#FFC107" {
		t.Errorf("Colors.Amber = %s, want #FFC107", got)
	}

	if got := Colors.Blue.String(); got != "#2196F3" {
		t.Errorf("Colors.Blue = %s, want #2196F3", got)
	}

	if Colors.Transparent.A != 0 {
		t.Errorf("Colors.Transparent.A = %d, want 0 (fully transparent)", Colors.Transparent.A)
	}
}

func TestTextScale(t *testing.T) {
	if TextSize.DisplayLarge != 57 {
		t.Errorf("TextSize.DisplayLarge = %v, want 57", TextSize.DisplayLarge)
	}

	if TextSize.HeadlineMedium != 28 {
		t.Errorf("TextSize.HeadlineMedium = %v, want 28", TextSize.HeadlineMedium)
	}

	if TextSize.BodyMedium != 14 {
		t.Errorf("TextSize.BodyMedium = %v, want 14", TextSize.BodyMedium)
	}
}
