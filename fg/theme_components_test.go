package fg

import (
	"testing"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// cardProps walks a Card widget and decodes its marshaled CardProps, so tests
// can assert on the value actually sent over the wire (not just the struct
// field, which the theme fallback only applies at walkNodes time).
func cardProps(t *testing.T, c *CardWidget) *fugov1.CardProps {
	t.Helper()

	var counter uint32

	nodes := c.walkNodes(&counter)
	if len(nodes) == 0 {
		t.Fatal("walkNodes returned no nodes")
	}

	props := &fugov1.CardProps{}
	if err := proto.Unmarshal(nodes[0].GetProps(), props); err != nil {
		t.Fatalf("unmarshal CardProps: %v", err)
	}

	return props
}

func TestCardUsesThemeComponentDefaultsWhenUnset(t *testing.T) {
	defer UseTheme(LightTheme())

	theme := LightTheme()
	theme.Components.CardRadius = 20
	theme.Components.CardElevation = 3
	UseTheme(theme)

	props := cardProps(t, Card(Text("hi")))

	if got, want := props.GetBorderRadius(), 20.0; got != want {
		t.Errorf("Card BorderRadius = %v, want theme default %v", got, want)
	}

	if got, want := props.GetElevation(), 3.0; got != want {
		t.Errorf("Card Elevation = %v, want theme default %v", got, want)
	}
}

func TestCardExplicitValuesOverrideTheme(t *testing.T) {
	defer UseTheme(LightTheme())

	theme := LightTheme()
	theme.Components.CardRadius = 20
	theme.Components.CardElevation = 3
	UseTheme(theme)

	props := cardProps(t, Card(Text("hi")).BorderRadius(2).Elevation(5))

	if got, want := props.GetBorderRadius(), 2.0; got != want {
		t.Errorf("explicit Card BorderRadius = %v, want %v (theme must not override)", got, want)
	}

	if got, want := props.GetElevation(), 5.0; got != want {
		t.Errorf("explicit Card Elevation = %v, want %v (theme must not override)", got, want)
	}
}

func TestButtonUsesThemeComponentDefaultWhenUnset(t *testing.T) {
	defer UseTheme(LightTheme())

	theme := LightTheme()
	theme.Components.ButtonRadius = 24
	UseTheme(theme)

	b := Button("x")

	var counter uint32

	nodes := b.walkNodes(&counter)

	props := &fugov1.ButtonProps{}
	if err := proto.Unmarshal(nodes[0].GetProps(), props); err != nil {
		t.Fatalf("unmarshal ButtonProps: %v", err)
	}

	if got, want := props.GetBorderRadius(), 24.0; got != want {
		t.Errorf("Button BorderRadius = %v, want theme default %v", got, want)
	}
}

func TestButtonExplicitBorderRadiusOverridesTheme(t *testing.T) {
	defer UseTheme(LightTheme())

	theme := LightTheme()
	theme.Components.ButtonRadius = 24
	UseTheme(theme)

	b := Button("x").BorderRadius(6)

	var counter uint32

	nodes := b.walkNodes(&counter)

	props := &fugov1.ButtonProps{}
	if err := proto.Unmarshal(nodes[0].GetProps(), props); err != nil {
		t.Fatalf("unmarshal ButtonProps: %v", err)
	}

	if got, want := props.GetBorderRadius(), 6.0; got != want {
		t.Errorf("explicit Button BorderRadius = %v, want %v (theme must not override)", got, want)
	}
}
