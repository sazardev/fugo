//nolint:ireturn
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/style"
	"github.com/sazardev/fugo/supervisor"
	"github.com/sazardev/fugo/transport"
	"github.com/sazardev/fugo/ui"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

const (
	defaultAddr     = "127.0.0.1:9510"
	defaultWidth    = 800
	defaultHeight   = 600
	shutdownTimeout = 5 * time.Second
)

func main() {
	app := fugo.NewApp(fugo.AppOptions{
		Title:  "Fugo Demo - Router + Fluent API",
		Width:  defaultWidth,
		Height: defaultHeight,
	})

	server, _, err := transport.StartServer(defaultAddr, app)
	if err != nil {
		log.Fatalf("start server: %v", err)
	}

	flutterBinary := findFlutterBinary()

	proc, err := supervisor.StartFlutter(context.Background(), defaultAddr, flutterBinary)
	if err != nil {
		server.GracefulStop()
		log.Fatalf("start flutter: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigCh:
			log.Println("[fugo] signal received")
		case <-proc.Exited():
			log.Println("[fugo] flutter window closed")
		}
		log.Println("[fugo] shutting down")
		app.Shutdown()
		if err := proc.Shutdown(shutdownTimeout); err != nil {
			log.Printf("[fugo] shutdown error: %v", err)
		}
		server.GracefulStop()
		os.Exit(0)
	}()

	log.Println("[fugo] starting app")
	app.Run(buildUI)
}

func buildUI(ctx *fugo.Context) ui.Widget {
	return ui.NewRouter(map[string]func() ui.Widget{
		"/":        func() ui.Widget { return homePage(ctx) },
		"/inputs":  func() ui.Widget { return inputsPage(ctx) },
		"/gallery": func() ui.Widget { return galleryPage(ctx) },
	}, "/")
}

// --- Home Page ---
func homePage(ctx *fugo.Context) ui.Widget {
	accent := style.Hex("#0F3460")
	green := style.Hex("#10B981")

	counter := 0
	counterText := ui.NewText("0").
		WithFontSize(32).
		WithColor(style.Hex("#FFFFFF"))

	decBtn := ui.NewButton("-").
		WithBgColor(style.Hex("#EF4444")).
		OnClick(func(_ ui.Event) {
			counter--
			counterText.SetText(strconv.Itoa(counter))
			ctx.Update()
		})

	incBtn := ui.NewButton("+").
		WithBgColor(green).
		OnClick(func(_ ui.Event) {
			counter++
			counterText.SetText(strconv.Itoa(counter))
			ctx.Update()
		})

	counterRow := ui.NewRow(decBtn, counterText, incBtn).
		WithMainAlign(fugov1.MainAxisAlignment_MAIN_CENTER)

	goInputs := ui.NewButton("Go to Inputs →").
		WithBgColor(accent).
		OnClick(func(_ ui.Event) {
			ctx.NavigateTo("/inputs")
		})

	goGallery := ui.NewButton("Go to Gallery →").
		WithBgColor(style.Hex("#8B5CF6")).
		OnClick(func(_ ui.Event) {
			ctx.NavigateTo("/gallery")
		})

	header := ui.NewText("Home").WithFontSize(20).WithWeight(style.WeightBold)

	return ui.NewContainer(
		ui.NewColumn(
			ui.NewPadding(header, 0, 0, 12, 0),
			counterRow,
			ui.NewSizedBox(0, 24),
			goInputs,
			ui.NewSizedBox(0, 8),
			goGallery,
		),
	).WithBgColor(style.Hex("#1A1A2E")).WithPad(style.EdgeAll(24))
}

// --- Inputs Page ---
func inputsPage(ctx *fugo.Context) ui.Widget {
	accent := style.Hex("#0F3460")

	checkboxOn := false
	switchOn := false
	sliderVal := 50.0

	checkboxStatus := ui.NewText("Checkbox: OFF").
		WithColor(style.Hex("#9CA3AF"))
	switchStatus := ui.NewText("Switch: OFF").
		WithColor(style.Hex("#9CA3AF"))
	sliderText := ui.NewText("Slider: 50").
		WithColor(style.Hex("#9CA3AF"))
	textfieldEcho := ui.NewText("").
		WithColor(style.Hex("#9CA3AF"))
	counter := 0
	animStatus := ui.NewText("Tap button to cycle color").
		WithColor(style.Hex("#9CA3AF"))

	tf := ui.NewTextField("Type something...").
		OnChange(func(e ui.Event) {
			textfieldEcho.SetText(string(e.Data))
			ctx.Update()
		})

	cb := ui.NewCheckbox("Toggle me").
		OnChange(func(_ ui.Event) {
			checkboxOn = !checkboxOn
			if checkboxOn {
				checkboxStatus.SetText("Checkbox: ON")
			} else {
				checkboxStatus.SetText("Checkbox: OFF")
			}
			ctx.Update()
		})

	sw := ui.NewSwitch().
		OnChange(func(_ ui.Event) {
			switchOn = !switchOn
			if switchOn {
				switchStatus.SetText("Switch: ON")
			} else {
				switchStatus.SetText("Switch: OFF")
			}
			ctx.Update()
		})

	sl := ui.NewSlider().
		WithMin(0).WithMax(100).
		WithValue(sliderVal)
	sl.OnChange(func(e ui.Event) {
		if v, err := strconv.ParseFloat(string(e.Data), 64); err == nil {
			sliderVal = v
			sl.Value = v
			sliderText.SetText("Slider: " + strconv.Itoa(int(v)))
			ctx.Update()
		}
	})

	animBg := style.Hex("#0F3460")
	animCont := ui.NewAnimatedContainer(
		ui.NewPadding(ui.NewText("Tap to animate"), 12, 12, 12, 12),
	).WithBgColor(animBg).WithDurationMs(300)

	animContBtn := ui.NewButton("Cycle color").
		WithBgColor(style.Hex("#F59E0B")).
		OnClick(func(_ ui.Event) {
			colors := []style.Color{
				style.Hex("#0F3460"), style.Hex("#10B981"),
				style.Hex("#8B5CF6"), style.Hex("#F59E0B"),
				style.Hex("#EF4444"),
			}
			animBg = colors[counter%len(colors)]
			animCont.BgColor = animBg
			counter++
			animStatus.SetText("Cycled " + strconv.Itoa(counter) + " times")
			ctx.Update()
		})

	backBtn := ui.NewButton("← Back").
		WithBgColor(accent).
		OnClick(func(_ ui.Event) {
			ctx.GoBack()
		})

	header := ui.NewText("Inputs").WithFontSize(20).WithWeight(style.WeightBold)

	sectionLabel := func(s string) ui.Widget {
		return ui.NewText(s).WithFontSize(14).WithWeight(style.WeightBold).WithColor(style.Hex("#FFFFFF"))
	}

	return ui.NewContainer(
		ui.NewColumn(
			ui.NewRow(backBtn, ui.NewSizedBox(16, 0), header),
			ui.NewSizedBox(0, 16),

			sectionLabel("TextField"),
			tf,
			ui.NewSizedBox(0, 4),
			textfieldEcho,
			ui.NewSizedBox(0, 12),

			sectionLabel("Checkbox + Switch"),
			cb,
			checkboxStatus,
			ui.NewSizedBox(0, 4),
			sw,
			switchStatus,
			ui.NewSizedBox(0, 12),

			sectionLabel("Slider"),
			sl,
			sliderText,
			ui.NewSizedBox(0, 12),

			sectionLabel("AnimatedContainer"),
			animCont,
			ui.NewSizedBox(0, 4),
			animContBtn,
			ui.NewSizedBox(0, 4),
			animStatus,
		),
	).WithBgColor(style.Hex("#1A1A2E")).WithPad(style.EdgeAll(24))
}

// --- Gallery Page ---
func galleryPage(ctx *fugo.Context) ui.Widget {
	accent := style.Hex("#0F3460")
	red := style.Hex("#EF4444")
	green := style.Hex("#10B981")
	blue := style.Hex("#3B82F6")
	orange := style.Hex("#F59E0B")
	purple := style.Hex("#8B5CF6")
	pink := style.Hex("#EC4899")

	// Row + Expanded
	redBox := ui.NewContainer(ui.NewPadding(ui.NewText("Red"), 8, 8, 8, 8)).WithBgColor(red)
	greenBox := ui.NewContainer(ui.NewPadding(ui.NewText("Green"), 8, 8, 8, 8)).WithBgColor(green)
	blueBox := ui.NewContainer(ui.NewPadding(ui.NewText("Blue"), 8, 8, 8, 8)).WithBgColor(blue)

	boxRow := ui.NewRow(
		ui.NewExpanded(redBox),
		ui.NewExpanded(greenBox),
		ui.NewExpanded(blueBox),
	)

	// Stack + Positioned
	stackBg := ui.NewContainer(ui.NewSizedBox(250, 80)).WithBgColor(accent)
	stack := ui.NewStack(
		stackBg,
		ui.NewPositioned(ui.NewText("TL")).WithLeft(8).WithTop(8),
		ui.NewPositioned(ui.NewText("TR")).WithRight(8).WithTop(8),
		ui.NewPositioned(ui.NewText("C")).WithLeft(120).WithTop(30),
		ui.NewPositioned(ui.NewText("BL")).WithLeft(8).WithBottom(8),
		ui.NewPositioned(ui.NewText("BR")).WithRight(8).WithBottom(8),
	)

	// GridView
	gridColors := []style.Color{red, orange, green, blue, purple, pink}
	var gridItems []ui.Widget
	for _, c := range gridColors {
		gridItems = append(
			gridItems,
			ui.NewContainer(ui.NewSizedBox(50, 50)).WithBgColor(c),
		)
	}
	gridView := ui.NewGridView(gridItems...).
		WithCrossAxisCount(3).WithChildAspectRatio(1.5)
	grid := ui.NewSizedBox(0, 120).WithChild(gridView)

	// Wrap
	var wrapItems []ui.Widget
	for _, c := range []style.Color{red, blue, green, orange, purple} {
		wrapItems = append(
			wrapItems,
			ui.NewContainer(ui.NewPadding(ui.NewText("chip"), 4, 4, 4, 4)).WithBgColor(c),
		)
	}
	wrap := ui.NewWrap(wrapItems...).WithSpacing(6).WithRunSpacing(4)

	// Icons
	horizGap := ui.NewSizedBox(12, 0)
	iconRow := ui.NewRow(
		ui.NewIcon("home"), horizGap,
		ui.NewIcon("star"), horizGap,
		ui.NewIcon("favorite"), horizGap,
		ui.NewIcon("settings"), horizGap,
		ui.NewIcon("info"),
	).WithMainAlign(fugov1.MainAxisAlignment_MAIN_CENTER)

	divider := ui.NewDivider().WithColor(style.Hex("#6B7280")).WithThickness(1)

	backBtn := ui.NewButton("← Back").
		WithBgColor(accent).
		OnClick(func(_ ui.Event) {
			ctx.GoBack()
		})

	header := ui.NewText("Gallery").WithFontSize(20).WithWeight(style.WeightBold)

	sectionLabel := func(s string) ui.Widget {
		return ui.NewText(s).WithFontSize(14).WithWeight(style.WeightBold).WithColor(style.Hex("#FFFFFF"))
	}

	return ui.NewContainer(
		ui.NewColumn(
			ui.NewRow(backBtn, ui.NewSizedBox(16, 0), header),
			ui.NewSizedBox(0, 16),

			sectionLabel("Row + Expanded"),
			boxRow,
			ui.NewSizedBox(0, 16),

			sectionLabel("Stack + Positioned"),
			stack,
			ui.NewSizedBox(0, 16),

			sectionLabel("GridView (in SizedBox)"),
			grid,
			ui.NewSizedBox(0, 16),

			sectionLabel("Wrap"),
			wrap,
			ui.NewSizedBox(0, 16),

			sectionLabel("Icons + Divider"),
			iconRow,
			ui.NewSizedBox(0, 8),
			divider,
		),
	).WithBgColor(style.Hex("#1A1A2E")).WithPad(style.EdgeAll(24))
}

func findFlutterBinary() string {
	candidates := []string{
		"flutter_client/build/windows/x64/runner/Release/fugo_flutter_client.exe",
		"flutter_client/build/linux/x64/debug/bundle/fugo_flutter_client",
		"flutter_client/build/linux/x64/release/bundle/fugo_flutter_client",
		os.Getenv("FUGO_FLUTTER_BINARY"),
	}
	for _, path := range candidates {
		if path == "" {
			continue
		}

		cleanPath := filepath.Clean(path)
		if _, err := os.Stat(cleanPath); err == nil {
			return cleanPath
		}
	}

	log.Fatal("flutter binary not found. Run: cd flutter_client && flutter build windows")

	return ""
}
