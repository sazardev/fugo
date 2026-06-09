package main

import "fmt"

// projectFiles is the generated source for a chosen starter template: the
// entrypoint theme, the ui-package body, and the default window size written
// into fugo.toml. The window title and the import path are filled in per call.
type projectFiles struct {
	theme  string // "Light" | "Dark" — the fg.<theme>Theme() set in main.go
	uiHome string // ui/home.go source
	width  int
	height int
}

// filesFor builds the per-template sources, interpolating the project title
// where a template needs it.
func filesFor(template, title string) projectFiles {
	switch template {
	case "app":
		return projectFiles{theme: "Dark", uiHome: appUI, width: 900, height: 640}
	case "showcase":
		return projectFiles{theme: "Dark", uiHome: showcaseUI, width: 980, height: 760}
	default:
		return projectFiles{theme: "Light", uiHome: fmt.Sprintf(counterUI, title), width: 800, height: 600}
	}
}

// mainGo renders main.go for a module path and theme ("Light"/"Dark").
func mainGo(module, theme string) string {
	return fmt.Sprintf(mainTemplate, theme, module)
}

// mainTemplate is the thin entrypoint shared by every template: set the theme,
// then run the ui package's Build with options loaded from fugo.toml.
// %[1]s = theme word, %[2]s = module path.
const mainTemplate = `// Command %[2]s is a Fugo app: Go owns all logic, state and routing; a
// precompiled Flutter binary renders the UI over a local gRPC stream.
package main

import (
	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"

	"%[2]s/ui"
)

func main() {
	// Theme the whole app; try fg.DarkTheme() (or edit fg colors).
	fg.UseTheme(fg.%[1]sTheme())

	// Window title/size and the gRPC address are read from fugo.toml.
	fugo.RunStandalone(fugo.ConfigOptions("fugo.toml"), ui.Build)
}
`

// counterUI is the default template: a centered counter with two FABs.
// %[1]s = AppBar title.
const counterUI = `// Package ui holds the app's screens.
package ui

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

// Build returns the root widget. Fugo calls it once and retains the tree;
// event handlers mutate widgets in place and call ctx.Update() to re-render.
func Build(ctx *fugo.Context) fg.Widget {
	count := 0
	display := fg.Text("0").FontSize(57)

	update := func() {
		display.SetText(strconv.Itoa(count))
		ctx.Update()
	}

	return fg.Scaffold(
		fg.Center(display),
	).AppBar(fg.AppBar("%[1]s")).FAB(
		fg.Row(
			fg.FloatingActionButton("remove").OnClick(func(_ fg.Event) {
				count--
				update()
			}),
			fg.SizedBox(16, 0),
			fg.FloatingActionButton("add").OnClick(func(_ fg.Event) {
				count++
				update()
			}),
		),
	)
}
`

// appUI is the multi-page starter: a Router with Home and About pages.
const appUI = `// Package ui holds the app's screens.
package ui

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

// Build is the app's root: a Router with Home and About pages.
func Build(ctx *fugo.Context) fg.Widget {
	return fg.Router(map[string]func() fg.Widget{
		"/":      func() fg.Widget { return homePage(ctx) },
		"/about": func() fg.Widget { return aboutPage(ctx) },
	}, "/")
}

func homePage(ctx *fugo.Context) fg.Widget {
	t := fg.CurrentTheme()
	counter := 0
	count := fg.Text("0").FontSize(t.Typography.Heading * 2)

	inc := fg.Button("Increment").
		BgColor(t.Colors.Primary).
		OnClick(func(_ fg.Event) {
			counter++
			count.SetText(strconv.Itoa(counter))
			ctx.Update()
		})

	about := fg.Button("About →").
		BgColor(t.Colors.Secondary).
		OnClick(func(_ fg.Event) { ctx.NavigateTo("/about") })

	return page(t, "Home", count, fg.SizedBox(0, t.Spacing.MD), inc, fg.SizedBox(0, t.Spacing.SM), about)
}

func aboutPage(ctx *fugo.Context) fg.Widget {
	t := fg.CurrentTheme()
	back := fg.Button("← Back").
		BgColor(t.Colors.Surface).
		OnClick(func(_ fg.Event) { ctx.GoBack() })

	return page(t, "About",
		fg.Text("Built with Fugo — Go drives logic, Flutter renders.").Color(t.Colors.Muted),
		fg.SizedBox(0, t.Spacing.MD),
		back,
	)
}

func page(t fg.Theme, title string, body ...fg.Widget) fg.Widget {
	items := []fg.Widget{
		fg.Text(title).FontSize(t.Typography.Heading).Weight(fg.WeightBold),
		fg.Divider().Color(t.Colors.Border),
		fg.SizedBox(0, t.Spacing.MD),
	}
	items = append(items, body...)

	return fg.Container(fg.Column(items...)).
		BgColor(t.Colors.Background).
		Pad(fg.EdgeAll(t.Spacing.LG))
}
`

// showcaseUI puts most widgets on one scrollable page — a living reference.
const showcaseUI = `// Package ui holds the app's screens.
package ui

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

// Build is the showcase root: a Router with the gallery and an About page.
func Build(ctx *fugo.Context) fg.Widget {
	return fg.Router(map[string]func() fg.Widget{
		"/":      func() fg.Widget { return showcasePage(ctx) },
		"/about": func() fg.Widget { return aboutPage(ctx) },
	}, "/")
}

func showcasePage(ctx *fugo.Context) fg.Widget {
	t := fg.CurrentTheme()

	counter := 0
	countText := fg.Text("0").FontSize(t.Typography.Heading).Weight(fg.WeightBold)
	dec := fg.Button("−").BgColor(t.Colors.Error).OnClick(func(_ fg.Event) {
		counter--
		countText.SetText(strconv.Itoa(counter))
		ctx.Update()
	})
	inc := fg.Button("+").BgColor(t.Colors.Success).OnClick(func(_ fg.Event) {
		counter++
		countText.SetText(strconv.Itoa(counter))
		ctx.Update()
	})

	cbStatus := fg.Text("off").Color(t.Colors.Muted)
	cb := fg.Checkbox("Enable feature").OnChange(func(e fg.Event) {
		cbStatus.SetText(map[bool]string{true: "on", false: "off"}[string(e.Data) == "1"])
		ctx.Update()
	})
	swStatus := fg.Text("off").Color(t.Colors.Muted)
	sw := fg.Switch().OnChange(func(e fg.Event) {
		swStatus.SetText(map[bool]string{true: "on", false: "off"}[string(e.Data) == "1"])
		ctx.Update()
	})
	sliderText := fg.Text("50").Color(t.Colors.Muted)
	sl := fg.Slider().SetMin(0).SetMax(100).SetValue(50)
	sl.OnChange(func(e fg.Event) {
		if v, err := strconv.ParseFloat(string(e.Data), 64); err == nil {
			sl.SetValue(v)
			sliderText.SetText(strconv.Itoa(int(v)))
			ctx.Update()
		}
	})
	echo := fg.Text("…").Color(t.Colors.Muted)
	tf := fg.TextField("Type here").OnChange(func(e fg.Event) {
		echo.SetText(string(e.Data))
		ctx.Update()
	})

	pick := fg.Text("none").Color(t.Colors.Muted)
	radioA := fg.Radio("a", "Option A").Group("a")
	radioB := fg.Radio("b", "Option B").Group("a")
	radioA.OnChange(func(_ fg.Event) { radioA.GroupValue = "a"; radioB.GroupValue = "a"; pick.SetText("A"); ctx.Update() })
	radioB.OnChange(func(_ fg.Event) { radioA.GroupValue = "b"; radioB.GroupValue = "b"; pick.SetText("B"); ctx.Update() })
	dd := fg.Dropdown([]string{"Red", "Green", "Blue"}).SetValue("Red")
	dd.OnChange(func(e fg.Event) { dd.SetValue(string(e.Data)); pick.SetText(string(e.Data)); ctx.Update() })

	animColors := []fg.Color{t.Colors.Primary, t.Colors.Secondary, t.Colors.Success, t.Colors.Error}
	animIdx := 0
	anim := fg.AnimatedContainer(fg.PaddingAll(fg.Text("Tap me"), 16)).BgColor(animColors[0]).DurationMs(300)
	tap := fg.GestureDetector(anim).OnTap(func(_ fg.Event) {
		animIdx = (animIdx + 1) % len(animColors)
		anim.BgColor(animColors[animIdx])
		ctx.Update()
	})

	var tiles []fg.Widget
	for _, c := range []fg.Color{t.Colors.Primary, t.Colors.Secondary, t.Colors.Success, t.Colors.Error, fg.Hex("#F59E0B"), fg.Hex("#EC4899")} {
		tiles = append(tiles, fg.Container(fg.SizedBox(56, 56)).BgColor(c).BorderRadius(8))
	}
	grid := fg.SizedBox(0, 140).Child(fg.GridView(tiles...).CrossAxisCount(6).ChildAspectRatio(1))

	var chips []fg.Widget
	for _, s := range []string{"go", "flutter", "grpc", "protobuf", "impeller"} {
		chips = append(chips, fg.Container(fg.PaddingAll(fg.Text(s), 6)).BgColor(t.Colors.Surface).BorderRadius(12))
	}
	wrap := fg.Wrap(chips...).Spacing(8).RunSpacing(8)

	icons := fg.Row(
		fg.Icon("home").Size(28), fg.SizedBox(16, 0),
		fg.Icon("star").Size(28).Color(t.Colors.Success), fg.SizedBox(16, 0),
		fg.Icon("favorite").Size(28).Color(t.Colors.Error), fg.SizedBox(16, 0),
		fg.Icon("settings").Size(28),
	)

	body := fg.Column(
		fg.Text("Fugo Showcase").FontSize(t.Typography.Heading*1.6).Weight(fg.WeightBold),
		fg.Text("Go drives logic & state; Flutter renders. Dark theme, live.").Color(t.Colors.Muted),
		fg.SizedBox(0, t.Spacing.LG),

		card(t, "Buttons & counter", fg.Row(dec, fg.SizedBox(16, 0), countText, fg.SizedBox(16, 0), inc)),
		card(t, "Inputs",
			fg.Row(cb, fg.SizedBox(8, 0), cbStatus),
			fg.SizedBox(0, t.Spacing.SM),
			fg.Row(sw, fg.SizedBox(8, 0), swStatus),
			fg.SizedBox(0, t.Spacing.SM),
			sl, sliderText,
			fg.SizedBox(0, t.Spacing.SM),
			tf, echo,
		),
		card(t, "Selection", radioA, radioB, fg.SizedBox(0, t.Spacing.SM), dd,
			fg.SizedBox(0, t.Spacing.SM), fg.Row(fg.Text("Picked: ").Color(t.Colors.Muted), pick)),
		card(t, "Animation + gestures", tap),
		card(t, "Gallery (grid + wrap)", grid, fg.SizedBox(0, t.Spacing.SM), wrap),
		card(t, "Icons", icons),

		fg.SizedBox(0, t.Spacing.MD),
		fg.Button("About →").BgColor(t.Colors.Primary).OnClick(func(_ fg.Event) { ctx.NavigateTo("/about") }),
		fg.SizedBox(0, t.Spacing.XL),
	)

	return fg.Container(fg.ScrollView(body)).BgColor(t.Colors.Background).Pad(fg.EdgeAll(t.Spacing.LG))
}

func aboutPage(ctx *fugo.Context) fg.Widget {
	t := fg.CurrentTheme()

	return fg.Container(
		fg.Center(fg.Column(
			fg.Text("About").FontSize(t.Typography.Heading).Weight(fg.WeightBold),
			fg.SizedBox(0, t.Spacing.MD),
			fg.Text("Built with Fugo — one Go binary, native Flutter rendering.").Color(t.Colors.Muted),
			fg.SizedBox(0, t.Spacing.LG),
			fg.Button("← Back").BgColor(t.Colors.Surface).OnClick(func(_ fg.Event) { ctx.GoBack() }),
		)),
	).BgColor(t.Colors.Background).Pad(fg.EdgeAll(t.Spacing.LG))
}

func card(t fg.Theme, title string, body ...fg.Widget) fg.Widget {
	items := []fg.Widget{
		fg.Text(title).Weight(fg.WeightBold).Color(t.Colors.OnSurface),
		fg.Divider().Color(t.Colors.Border),
		fg.SizedBox(0, t.Spacing.SM),
	}
	items = append(items, body...)

	return fg.Container(fg.PaddingAll(fg.Column(items...), 16)).
		BgColor(t.Colors.Surface).BorderRadius(12)
}
`

// configTemplate is the generated fugo.toml.
// %[1]s = module/name, %[2]s = window title, %[3]d/%[4]d = width/height.
const configTemplate = `# Fugo project configuration.
# Read by 'fugo run' / 'fugo build' and by the app at startup.
name = "%[1]s"

[window]
title  = "%[2]s"
width  = %[3]d
height = %[4]d

[server]
# gRPC address the Go server listens on and the Flutter client dials.
addr = "127.0.0.1:9510"
`

// gitignoreTemplate is the generated .gitignore.
const gitignoreTemplate = `# Build output
/bin/
/dist/

# Runtime logs
/logs/*
!/logs/.gitkeep

# Go
*.test
*.out
go.work
go.work.sum

# OS / editors
.DS_Store
Thumbs.db
.idea/
.vscode/
`

// readmeTemplate is the generated README.md. %[1]s = project name.
const readmeTemplate = "# %[1]s\n\n" +
	"A desktop app built with [Fugo](https://github.com/sazardev/fugo) — you write\n" +
	"all logic, state and routing in **Go**; a precompiled Flutter binary renders the\n" +
	"UI over a local gRPC stream.\n\n" +
	"## Develop\n\n" +
	"```sh\n" +
	"fugo run            # build + launch (add --watch for hot reload)\n" +
	"```\n\n" +
	"## Ship\n\n" +
	"```sh\n" +
	"fugo build          # outputs dist/ (app binary + bundled Flutter client)\n" +
	"```\n\n" +
	"## Layout\n\n" +
	"```\n" +
	"%[1]s/\n" +
	"├─ main.go        # entrypoint: theme + ui.Build\n" +
	"├─ ui/            # your screens (package ui)\n" +
	"│  └─ home.go     # ui.Build — the root widget\n" +
	"├─ fugo.toml      # window title/size + gRPC address\n" +
	"├─ bin/           # dev builds        (gitignored)\n" +
	"├─ dist/          # release bundle    (gitignored)\n" +
	"└─ logs/          # last run's logs   (gitignored)\n" +
	"```\n\n" +
	"## Configure\n\n" +
	"Edit `fugo.toml` to change the window title, size, or gRPC address.\n"
