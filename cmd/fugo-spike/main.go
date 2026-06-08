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
	"github.com/sazardev/fugo/fg"
	"github.com/sazardev/fugo/supervisor"
	"github.com/sazardev/fugo/transport"

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

func buildUI(ctx *fugo.Context) fg.Widget {
	return fg.Router(map[string]func() fg.Widget{
		"/":        func() fg.Widget { return homePage(ctx) },
		"/inputs":  func() fg.Widget { return inputsPage(ctx) },
		"/gallery": func() fg.Widget { return galleryPage(ctx) },
	}, "/")
}

// Home page.
func homePage(ctx *fugo.Context) fg.Widget {
	accent := fg.Hex("#0F3460")
	green := fg.Hex("#10B981")

	counter := 0
	counterText := fg.Text("0").
		FontSize(32).
		Color(fg.Hex("#FFFFFF"))

	decBtn := fg.Button("-").
		BgColor(fg.Hex("#EF4444")).
		OnClick(func(_ fg.Event) {
			counter--
			counterText.SetText(strconv.Itoa(counter))
			ctx.Update()
		})

	incBtn := fg.Button("+").
		BgColor(green).
		OnClick(func(_ fg.Event) {
			counter++
			counterText.SetText(strconv.Itoa(counter))
			ctx.Update()
		})

	counterRow := fg.Row(decBtn, counterText, incBtn).
		MainAlign(fugov1.MainAxisAlignment_MAIN_CENTER)

	goInputs := fg.Button("Go to Inputs \u2192").
		BgColor(accent).
		OnClick(func(_ fg.Event) {
			ctx.NavigateTo("/inputs")
		})

	goGallery := fg.Button("Go to Gallery \u2192").
		BgColor(fg.Hex("#8B5CF6")).
		OnClick(func(_ fg.Event) {
			ctx.NavigateTo("/gallery")
		})

	header := fg.Text("Home").FontSize(20).Weight(fg.WeightBold)

	return fg.Container(
		fg.Column(
			fg.Padding(header, fg.EdgeOnly(0, 0, 12, 0)),
			counterRow,
			fg.SizedBox(0, 24),
			goInputs,
			fg.SizedBox(0, 8),
			goGallery,
		),
	).BgColor(fg.Hex("#1A1A2E")).Pad(fg.EdgeAll(24))
}

// Inputs page.
func inputsPage(ctx *fugo.Context) fg.Widget {
	accent := fg.Hex("#0F3460")

	checkboxOn := false
	switchOn := false
	sliderVal := 50.0

	checkboxStatus := fg.Text("Checkbox: OFF").
		Color(fg.Hex("#9CA3AF"))
	switchStatus := fg.Text("Switch: OFF").
		Color(fg.Hex("#9CA3AF"))
	sliderText := fg.Text("Slider: 50").
		Color(fg.Hex("#9CA3AF"))
	textfieldEcho := fg.Text("").
		Color(fg.Hex("#9CA3AF"))
	counter := 0
	animStatus := fg.Text("Tap button to cycle color").
		Color(fg.Hex("#9CA3AF"))

	tf := fg.TextField("Type something...").
		OnChange(func(e fg.Event) {
			textfieldEcho.SetText(string(e.Data))
			ctx.Update()
		})

	cb := fg.Checkbox("Toggle me").
		OnChange(func(_ fg.Event) {
			checkboxOn = !checkboxOn
			if checkboxOn {
				checkboxStatus.SetText("Checkbox: ON")
			} else {
				checkboxStatus.SetText("Checkbox: OFF")
			}
			ctx.Update()
		})

	sw := fg.Switch().
		OnChange(func(_ fg.Event) {
			switchOn = !switchOn
			if switchOn {
				switchStatus.SetText("Switch: ON")
			} else {
				switchStatus.SetText("Switch: OFF")
			}
			ctx.Update()
		})

	sl := fg.Slider().
		SetMin(0).SetMax(100).
		SetValue(sliderVal)
	sl.OnChange(func(e fg.Event) {
		if v, err := strconv.ParseFloat(string(e.Data), 64); err == nil {
			sliderVal = v
			sl.Value = v
			sliderText.SetText("Slider: " + strconv.Itoa(int(v)))
			ctx.Update()
		}
	})

	animBg := fg.Hex("#0F3460")
	animCont := fg.AnimatedContainer(
		fg.PaddingAll(fg.Text("Tap to animate"), 12),
	).BgColor(animBg).DurationMs(300)

	animContBtn := fg.Button("Cycle color").
		BgColor(fg.Hex("#F59E0B")).
		OnClick(func(_ fg.Event) {
			colors := []fg.Color{
				fg.Hex("#0F3460"), fg.Hex("#10B981"),
				fg.Hex("#8B5CF6"), fg.Hex("#F59E0B"),
				fg.Hex("#EF4444"),
			}
			animBg = colors[counter%len(colors)]
			animCont.BgColor(animBg)
			counter++
			animStatus.SetText("Cycled " + strconv.Itoa(counter) + " times")
			ctx.Update()
		})

	backBtn := fg.Button("\u2190 Back").
		BgColor(accent).
		OnClick(func(_ fg.Event) {
			ctx.GoBack()
		})

	header := fg.Text("Inputs").FontSize(20).Weight(fg.WeightBold)

	sectionLabel := func(s string) fg.Widget {
		return fg.Text(s).FontSize(14).Weight(fg.WeightBold).Color(fg.Hex("#FFFFFF"))
	}

	return fg.Container(
		fg.Column(
			fg.Row(backBtn, fg.SizedBox(16, 0), header),
			fg.SizedBox(0, 16),

			sectionLabel("TextField"),
			tf,
			fg.SizedBox(0, 4),
			textfieldEcho,
			fg.SizedBox(0, 12),

			sectionLabel("Checkbox + Switch"),
			cb,
			checkboxStatus,
			fg.SizedBox(0, 4),
			sw,
			switchStatus,
			fg.SizedBox(0, 12),

			sectionLabel("Slider"),
			sl,
			sliderText,
			fg.SizedBox(0, 12),

			sectionLabel("AnimatedContainer"),
			animCont,
			fg.SizedBox(0, 4),
			animContBtn,
			fg.SizedBox(0, 4),
			animStatus,
		),
	).BgColor(fg.Hex("#1A1A2E")).Pad(fg.EdgeAll(24))
}

// Gallery page.
func galleryPage(ctx *fugo.Context) fg.Widget {
	accent := fg.Hex("#0F3460")
	red := fg.Hex("#EF4444")
	green := fg.Hex("#10B981")
	blue := fg.Hex("#3B82F6")
	orange := fg.Hex("#F59E0B")
	purple := fg.Hex("#8B5CF6")
	pink := fg.Hex("#EC4899")

	// Row + Expanded
	redBox := fg.Container(fg.PaddingAll(fg.Text("Red"), 8)).BgColor(red)
	greenBox := fg.Container(fg.PaddingAll(fg.Text("Green"), 8)).BgColor(green)
	blueBox := fg.Container(fg.PaddingAll(fg.Text("Blue"), 8)).BgColor(blue)

	boxRow := fg.Row(
		fg.Expanded(redBox),
		fg.Expanded(greenBox),
		fg.Expanded(blueBox),
	)

	// Stack + Positioned
	stackBg := fg.Container(fg.SizedBox(250, 80)).BgColor(accent)
	stack := fg.Stack(
		stackBg,
		fg.Positioned(fg.Text("TL")).Left(8).Top(8),
		fg.Positioned(fg.Text("TR")).Right(8).Top(8),
		fg.Positioned(fg.Text("C")).Left(120).Top(30),
		fg.Positioned(fg.Text("BL")).Left(8).Bottom(8),
		fg.Positioned(fg.Text("BR")).Right(8).Bottom(8),
	)

	// GridView
	gridColors := []fg.Color{red, orange, green, blue, purple, pink}
	gridItems := make([]fg.Widget, 0, len(gridColors))
	for _, c := range gridColors {
		gridItems = append(
			gridItems,
			fg.Container(fg.SizedBox(50, 50)).BgColor(c),
		)
	}
	gridView := fg.GridView(gridItems...).
		CrossAxisCount(3).ChildAspectRatio(1.5)
	grid := fg.SizedBox(0, 120).Child(gridView)

	// Wrap
	wrapColors := []fg.Color{red, blue, green, orange, purple}
	wrapItems := make([]fg.Widget, 0, len(wrapColors))
	for _, c := range wrapColors {
		wrapItems = append(
			wrapItems,
			fg.Container(fg.PaddingAll(fg.Text("chip"), 4)).BgColor(c),
		)
	}
	wrap := fg.Wrap(wrapItems...).Spacing(6).RunSpacing(4)

	// Icons
	horizGap := fg.SizedBox(12, 0)
	iconRow := fg.Row(
		fg.Icon("home"), horizGap,
		fg.Icon("star"), horizGap,
		fg.Icon("favorite"), horizGap,
		fg.Icon("settings"), horizGap,
		fg.Icon("info"),
	).MainAlign(fugov1.MainAxisAlignment_MAIN_CENTER)

	divider := fg.Divider().Color(fg.Hex("#6B7280")).Thickness(1)

	backBtn := fg.Button("\u2190 Back").
		BgColor(accent).
		OnClick(func(_ fg.Event) {
			ctx.GoBack()
		})

	header := fg.Text("Gallery").FontSize(20).Weight(fg.WeightBold)

	sectionLabel := func(s string) fg.Widget {
		return fg.Text(s).FontSize(14).Weight(fg.WeightBold).Color(fg.Hex("#FFFFFF"))
	}

	return fg.Container(
		fg.Column(
			fg.Row(backBtn, fg.SizedBox(16, 0), header),
			fg.SizedBox(0, 16),

			sectionLabel("Row + Expanded"),
			boxRow,
			fg.SizedBox(0, 16),

			sectionLabel("Stack + Positioned"),
			stack,
			fg.SizedBox(0, 16),

			sectionLabel("GridView (in SizedBox)"),
			grid,
			fg.SizedBox(0, 16),

			sectionLabel("Wrap"),
			wrap,
			fg.SizedBox(0, 16),

			sectionLabel("Icons + Divider"),
			iconRow,
			fg.SizedBox(0, 8),
			divider,
		),
	).BgColor(fg.Hex("#1A1A2E")).Pad(fg.EdgeAll(24))
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
