// Command fugo is the Fugo CLI: init, run, build and doctor.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

const (
	osWindows   = "windows"
	subcmdBuild = "build"
)

func main() {
	cmd := &cli.Command{
		Name:    "fugo",
		Usage:   "Go SDUI framework CLI",
		Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
		Commands: []*cli.Command{
			initCmd(),
			runCmd(),
			buildCmd(),
			doctorCmd(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// scaffoldMain returns the main.go source for the chosen starter template,
// with the project name interpolated into the window title.
func scaffoldMain(template, name string) string {
	switch template {
	case "app":
		return fmt.Sprintf(appTemplate, name)
	case "showcase":
		return fmt.Sprintf(showcaseTemplate, name)
	default:
		return fmt.Sprintf(counterTemplate, name)
	}
}

const counterTemplate = `package main

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

func main() {
	fugo.RunStandalone(fugo.AppOptions{
		Title:  "%s",
		Width:  800,
		Height: 600,
	}, buildUI)
}

func buildUI(ctx *fugo.Context) fg.Widget {
	counter := 0
	counterText := fg.Text("0").FontSize(48)

	incBtn := fg.Button("+").
		BgColor(fg.Hex("#10B981")).
		FontSize(20).
		OnClick(func(_ fg.Event) {
			counter++
			counterText.SetText(strconv.Itoa(counter))
			ctx.Update()
		})

	return fg.Container(
		fg.Column(
			counterText,
			fg.SizedBox(0, 16),
			incBtn,
		),
	).BgColor(fg.Hex("#1A1A2E")).Pad(fg.EdgeAll(24))
}
`

const appTemplate = `package main

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

func main() {
	fg.UseTheme(fg.DarkTheme()) // try fg.LightTheme() to re-skin the whole app

	fugo.RunStandalone(fugo.AppOptions{
		Title:  "%s",
		Width:  900,
		Height: 640,
	}, buildUI)
}

func buildUI(ctx *fugo.Context) fg.Widget {
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

const showcaseTemplate = `package main

import (
	"strconv"

	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

func main() {
	fg.UseTheme(fg.DarkTheme())

	fugo.RunStandalone(fugo.AppOptions{
		Title:  "%s",
		Width:  980,
		Height: 760,
	}, buildUI)
}

func buildUI(ctx *fugo.Context) fg.Widget {
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
		animIdx = (animIdx + 1) %% len(animColors)
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

func initCmd() *cli.Command {
	var (
		fugoSrc  string
		template string
	)

	return &cli.Command{
		Name:      "init",
		Usage:     "Create a new Fugo project",
		ArgsUsage: "<project-name>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "fugo-src",
				Destination: &fugoSrc,
				Usage:       "path to local fugo source (for replace directive)",
			},
			&cli.StringFlag{
				Name:        "template",
				Aliases:     []string{"t"},
				Value:       "counter",
				Destination: &template,
				Usage:       "starter template: counter | app | showcase",
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			name := c.Args().First()
			if name == "" {
				return errors.New("project name required: fugo init <name>")
			}

			dir := filepath.Clean(name)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}

			modulePath := name
			if !strings.Contains(name, "/") {
				modulePath = name
			}

			// Write main.go
			mainFile := filepath.Join(dir, "main.go")
			content := scaffoldMain(template, name)

			if err := os.WriteFile(mainFile, []byte(content), 0o644); err != nil {
				return fmt.Errorf("write main.go: %w", err)
			}

			// go mod init
			fmt.Println("Initializing Go module...")
			modInit := exec.CommandContext(ctx, "go", "mod", "init", modulePath)
			modInit.Dir = dir
			modInit.Stdout = os.Stdout
			modInit.Stderr = os.Stderr
			if err := modInit.Run(); err != nil {
				return fmt.Errorf("go mod init: %w", err)
			}

			// Find fugo source: --fugo-src flag > auto-detect from CWD/executable
			fugoDir := fugoSrc
			if fugoDir == "" {
				fugoDir = findFugoRepo()
			}
			if fugoDir != "" {
				absFugo, _ := filepath.Abs(fugoDir)
				absProject, _ := filepath.Abs(dir)
				relToProject, err := filepath.Rel(absProject, absFugo)
				if err != nil {
					relToProject = fugoDir
				}
				relPath := strings.ReplaceAll(relToProject, "\\", "/")

				fmt.Printf("Detected local fugo source, adding replace => %s\n", relPath)
				goModFile := filepath.Join(dir, "go.mod")
				data, err := os.ReadFile(goModFile)
				if err != nil {
					return fmt.Errorf("read go.mod: %w", err)
				}
				replaceLine := fmt.Sprintf("\nreplace github.com/sazardev/fugo => %s\n", relPath)
				data = append(data, []byte(replaceLine)...)
				if err := os.WriteFile(goModFile, data, 0o644); err != nil {
					return fmt.Errorf("write go.mod: %w", err)
				}
			}

			// go mod tidy
			fmt.Println("Resolving dependencies...")
			tidy := exec.CommandContext(ctx, "go", "mod", "tidy")
			tidy.Dir = dir
			tidy.Stdout = os.Stdout
			tidy.Stderr = os.Stderr
			if err := tidy.Run(); err != nil {
				return fmt.Errorf("go mod tidy: %w", err)
			}

			fmt.Printf("\nCreated %s/\n", dir)
			fmt.Printf("  main.go  — your Fugo app\n")
			fmt.Printf("  go.mod   — Go module\n")
			fmt.Printf("\nNext: cd %s && fugo run\n", dir)

			return nil
		},
	}
}

func findFugoRepo() string {
	// Search from CWD upward
	dir, _ := os.Getwd()
	if found := searchUpForFugo(dir); found != "" {
		return found
	}

	// Search from executable location upward
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		if found := searchUpForFugo(exeDir); found != "" {
			return found
		}
	}

	// Check common locations
	candidates := []string{
		filepath.Join(os.Getenv("USERPROFILE"), "Documents", "work", "fugo"),
		filepath.Join(os.Getenv("HOME"), "fugo"),
	}
	for _, path := range candidates {
		goMod := filepath.Join(path, "go.mod")
		if data, err := os.ReadFile(goMod); err == nil && strings.Contains(string(data), "github.com/sazardev/fugo") {
			return path
		}
	}

	return ""
}

func searchUpForFugo(start string) string {
	dir := start
	for {
		goMod := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(goMod)
		if err == nil && strings.Contains(string(data), "github.com/sazardev/fugo") {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func runCmd() *cli.Command {
	var (
		addr    string
		flutter string
		watch   bool
	)

	return &cli.Command{
		Name:  "run",
		Usage: "Run the Fugo app in current directory",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Value:       "127.0.0.1:9510",
				Destination: &addr,
				Usage:       "listen address for gRPC server",
			},
			&cli.StringFlag{
				Name:        "flutter",
				Destination: &flutter,
				Usage:       "path to Flutter binary (auto-detect if empty)",
			},
			&cli.BoolFlag{
				Name:        "watch",
				Aliases:     []string{"w"},
				Destination: &watch,
				Usage:       "watch .go files and restart on change",
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			if !hasMainGo() {
				return errors.New("no main.go found in current directory. Run 'fugo init <name>' first")
			}

			fmt.Printf("=== Fugo v%s ===\n", version)

			if flutter == "" {
				ensureFlutterClient(ctx)
			}

			if watch {
				return runWithWatch(ctx, addr, flutter)
			}

			return buildAndRun(ctx, addr, flutter)
		},
	}
}

func buildAndRun(ctx context.Context, addr, flutter string) error {
	fmt.Println("Building...")
	build := exec.CommandContext(ctx, "go", subcmdBuild, "-o", appBinary(), ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("Running... (press Ctrl+C to stop)")

	return runApp(ctx, addr, flutter)
}

func runWithWatch(ctx context.Context, addr, flutter string) error {
	fmt.Println("Hot reload: the window stays open across rebuilds; the Go server restarts on .go changes.")

	flutterProc, err := startFlutterClient(ctx, addr, flutter)
	if err != nil {
		fmt.Printf("Could not start the Flutter client (%v) — falling back to full restarts.\n", err)

		return runWithFullRestart(ctx, addr, flutter)
	}
	defer killProc(flutterProc)

	snap := fileSnapshot()
	for {
		fmt.Println("Building Go server...")
		if buildErr := buildApp(ctx); buildErr != nil {
			fmt.Printf("build failed: %v (waiting for changes)\n", buildErr)
			waitForChange(&snap)

			continue
		}

		server := startServerOnly(ctx, addr)
		waitForChange(&snap)
		killProc(server)
		fmt.Println("--- change detected, reloading Go server (window stays open) ---")
	}
}

func runWithFullRestart(ctx context.Context, addr, flutter string) error {
	fmt.Println("Watching .go files for changes...")

	snap := fileSnapshot()
	for {
		if err := buildAndRun(ctx, addr, flutter); err != nil {
			fmt.Printf("Run error: %v\n", err)
		}

		waitForChange(&snap)
		fmt.Println("--- File change detected, restarting ---")
	}
}

// startFlutterClient launches the Flutter render client once; it auto-reconnects
// when the Go server restarts, so the window survives hot reloads.
func startFlutterClient(ctx context.Context, addr, flutter string) (*exec.Cmd, error) {
	bin := flutter
	if bin == "" {
		dir := flutterBundleDir(ctx)
		if dir == "" {
			ensureFlutterClient(ctx)
			dir = flutterBundleDir(ctx)
		}
		if dir == "" {
			return nil, errors.New("flutter client not built")
		}
		bin = filepath.Join(dir, "fugo_flutter_client"+exeSuffix())
	}

	cmd := exec.CommandContext(ctx, bin)
	cmd.Env = append(os.Environ(), "FUGO_ADDR="+addr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd, cmd.Start()
}

func buildApp(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "go", subcmdBuild, "-o", appBinary(), ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// startServerOnly runs the built app in server-only mode (FUGO_NO_FLUTTER=1) so
// the externally-managed Flutter client reconnects to it across reloads.
func startServerOnly(ctx context.Context, addr string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, appBinary())
	cmd.Env = append(os.Environ(), "FUGO_ADDR="+addr, "FUGO_NO_FLUTTER=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Printf("start server: %v\n", err)

		return nil
	}

	return cmd
}

func waitForChange(snap *map[string]time.Time) {
	for {
		time.Sleep(500 * time.Millisecond)

		current := fileSnapshot()
		if !snapshotEq(*snap, current) {
			*snap = current

			return
		}
	}
}

func killProc(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func fileSnapshot() map[string]time.Time {
	snap := make(map[string]time.Time)

	_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // intentional fallback: skip unreadable paths during walk
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "bin" || base == "vendor" {
				return filepath.SkipDir
			}

			return nil
		}
		if filepath.Ext(path) == ".go" {
			snap[path] = info.ModTime()
		}

		return nil
	})

	return snap
}

func snapshotEq(a, b map[string]time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for k, t := range a {
		if !t.Equal(b[k]) {
			return false
		}
	}

	return true
}

func runApp(ctx context.Context, addr, flutter string) error {
	run := exec.CommandContext(ctx, appBinary())
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	run.Env = append(os.Environ(), "FUGO_ADDR="+addr)
	if flutter != "" {
		run.Env = append(run.Env, "FUGO_FLUTTER_BINARY="+flutter)
	}
	setNewProcessGroup(run)

	if err := run.Start(); err != nil {
		return fmt.Errorf("start app: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- run.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		_ = run.Process.Kill()
		<-done

		return ctx.Err()
	}
}

func buildCmd() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build the app and bundle the Flutter client into dist/",
		Action: func(ctx context.Context, _ *cli.Command) error {
			if !hasMainGo() {
				return errors.New("no main.go found in current directory. Run 'fugo init <name>' first")
			}

			outDir := "dist"
			appOut := filepath.Join(outDir, projectName()+exeSuffix())

			fmt.Println("Building app (release)...")
			build := exec.CommandContext(ctx, "go", subcmdBuild, "-ldflags=-s -w", "-o", appOut, ".")
			build.Stdout = os.Stdout
			build.Stderr = os.Stderr
			if err := build.Run(); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}

			src := flutterBundleDir(ctx)
			if src == "" {
				fmt.Println("  ! Flutter client bundle not found — built the Go binary only.")
				fmt.Println("    Build the client first: cd <fugo>/flutter_client && flutter build windows")
				fmt.Printf("\nBuild complete (app only): %s\n", appOut)

				return nil
			}

			fmt.Println("Bundling Flutter client...")
			if err := copyDir(src, filepath.Join(outDir, "flutter")); err != nil {
				return fmt.Errorf("bundle flutter client: %w", err)
			}

			fmt.Printf("\nBuild complete: %s/\n", outDir)
			fmt.Printf("  %s — your app\n", filepath.Base(appOut))
			fmt.Printf("  flutter/      — bundled render client\n")
			fmt.Printf("\nShip the whole %s/ folder; run it with: %s\n", outDir, appOut)

			return nil
		},
	}
}

// projectName returns the current directory's base name, used as the app binary name.
func projectName() string {
	dir, err := os.Getwd()
	if err != nil || dir == "" {
		return "app"
	}

	return filepath.Base(dir)
}

// exeSuffix is the executable extension for the host OS.
func exeSuffix() string {
	if runtime.GOOS == osWindows {
		return ".exe"
	}

	return ""
}

// fugoModuleDir resolves the on-disk directory of the fugo module via
// `go list -m`, honoring any replace directive; "" if it can't be resolved.
func fugoModuleDir(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Dir}}", "github.com/sazardev/fugo").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

// flutterBundleDir locates the precompiled Flutter client bundle inside the
// fugo module (resolved via fugoModuleDir so it honors a replace directive), or
// "" if fugo is not a local checkout or the client has not been built yet.
func flutterBundleDir(ctx context.Context) string {
	repo := fugoModuleDir(ctx)
	if repo == "" {
		return ""
	}

	for _, c := range []string{
		filepath.Join(repo, "flutter_client", "build", "windows", "x64", "runner", "Release"),
		filepath.Join(repo, "flutter_client", "build", "linux", "x64", "release", "bundle"),
		filepath.Join(repo, "flutter_client", "build", "linux", "x64", "debug", "bundle"),
	} {
		if fi, err := os.Stat(c); err == nil && fi.IsDir() {
			return c
		}
	}

	return ""
}

// ensureFlutterClient builds the Flutter render client once if it isn't built
// yet, so `fugo run` works without a manual `flutter build`. It is a no-op when
// the client is already built, flutter isn't on PATH, or the fugo source tree
// can't be located.
func ensureFlutterClient(ctx context.Context) {
	if flutterBundleDir(ctx) != "" {
		return
	}

	repo := fugoModuleDir(ctx)
	if repo == "" {
		return
	}

	clientDir := filepath.Join(repo, "flutter_client")
	if _, err := os.Stat(clientDir); err != nil {
		return
	}

	if _, err := exec.LookPath("flutter"); err != nil {
		fmt.Println("Flutter client not built and 'flutter' is not on PATH.")
		fmt.Println("Build it once with: cd flutter_client && flutter build windows")

		return
	}

	fmt.Println("Flutter client not built yet — building it once (this can take a few minutes)...")

	args := []string{subcmdBuild, "windows"}
	if runtime.GOOS != osWindows {
		args = []string{subcmdBuild, "linux", "--debug"}
	}

	cmd := exec.CommandContext(ctx, "flutter", args...)
	cmd.Dir = clientDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("flutter build failed: %v\n", err)
	}
}

// copyDir recursively copies the contents of src into dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}

		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		return copyFile(path, target)
	})
}

// copyFile copies a single file from src to dst (dst's parent must exist).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()

		return err
	}

	return out.Close()
}

func hasMainGo() bool {
	_, err := os.Stat("main.go")

	return err == nil
}

// appBinary returns the build output path for the current OS so that
// `fugo build` and `fugo run` always agree on the binary name.
func appBinary() string {
	if runtime.GOOS == osWindows {
		return "bin/app.exe"
	}

	return "bin/app"
}

func doctorCmd() *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "Check development environment",
		Action: func(ctx context.Context, _ *cli.Command) error {
			checks := []struct {
				name string
				cmd  string
				args []string
			}{
				{"Go", "go", []string{"version"}},
				{"Flutter", "flutter", []string{"--version"}},
				{"protoc", "protoc", []string{"--version"}},
				{"gofumpt", "gofumpt", []string{"-version"}},
			}

			fmt.Println("Fugo Doctor")
			fmt.Println("===========")
			fmt.Println()

			for _, check := range checks {
				cmd := exec.CommandContext(ctx, check.cmd, check.args...)
				out, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("  %-10s NOT FOUND\n", check.name)
				} else {
					firstLine := string(out)
					if idx := indexByte(firstLine, '\n'); idx >= 0 {
						firstLine = firstLine[:idx]
					}
					if idx := indexByte(firstLine, '\r'); idx >= 0 {
						firstLine = firstLine[:idx]
					}
					fmt.Printf("  %-10s %s\n", check.name, strings.TrimSpace(firstLine))
				}
			}

			fmt.Println()
			fmt.Printf("  OS/Arch   %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Println()

			return nil
		},
	}
}

func indexByte(s string, b byte) int {
	for i := range len(s) {
		if s[i] == b {
			return i
		}
	}

	return -1
}
