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
	"runtime/debug"
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
		Usage:   "Server-Driven UI framework for desktop — write Go, render with Flutter",
		Version: versionString(),
		Description: `Fugo lets you build native desktop apps writing only Go. Your logic, state
and routing run in a Go process; a precompiled Flutter binary renders the UI
over a local gRPC stream. Go is the single source of truth.

Typical workflow:
  fugo init myapp        scaffold a project (try --template app|showcase)
  cd myapp
  fugo run               build + launch the app (Go server + Flutter window)
  fugo run --watch       hot reload: rebuild on .go changes, window stays open
  fugo build             bundle a shippable dist/ (app + Flutter client)

Other commands:
  fugo widgets           browse the fg widget catalog and their doc comments
  fugo doctor            check your toolchain (Go, Flutter, protoc, gofumpt)

Every command accepts -V/--verbose (trace commands, paths, timings and the
app's runtime logs) and -q/--quiet (errors only). Colors honor NO_COLOR.`,
		Commands: []*cli.Command{
			initCmd(),
			runCmd(),
			buildCmd(),
			widgetsCmd(),
			doctorCmd(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		out.failf("%v", err)
		os.Exit(1)
	}
}

// versionString reports the CLI version. `make build` injects version, commit
// and date through -ldflags; a `go install` binary keeps the defaults, so we
// fall back to the module version and VCS stamps the Go toolchain embeds in the
// build info, keeping `fugo --version` accurate either way.
func versionString() string {
	v, c, d := version, commit, date

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
	}

	if v == "0.1.0" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = strings.TrimPrefix(info.Main.Version, "v")
	}

	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if c == "unknown" && len(s.Value) >= 7 {
				c = s.Value[:7]
			}
		case "vcs.time":
			if d == "unknown" && s.Value != "" {
				d = s.Value
			}
		}
	}

	return fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
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
	count := 0
	display := fg.Text("0").FontSize(57)

	update := func() {
		display.SetText(strconv.Itoa(count))
		ctx.Update()
	}

	return fg.Scaffold(
		fg.Center(display),
	).AppBar("Fugo").FAB(
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
		Description: `Scaffold a new Fugo project: write main.go, run 'go mod init', wire a replace
directive to your local fugo checkout (auto-detected), and run 'go mod tidy'.

Templates (--template, -t):
  counter   minimal counter — one screen, a button, live state (default)
  app       themed multi-page starter with a Router and Home/About pages
  showcase  every widget on one scrollable page — a living API reference

Examples:
  fugo init myapp
  fugo init myapp -t showcase
  fugo init myapp --fugo-src ../fugo`,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:        "fugo-src",
				Destination: &fugoSrc,
				Usage:       "path to a local fugo checkout for the go.mod replace directive (auto-detected if empty)",
			},
			&cli.StringFlag{
				Name:        "template",
				Aliases:     []string{"t"},
				Value:       "counter",
				Destination: &template,
				Usage:       "starter template: counter | app | showcase",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			setupUI()

			name := c.Args().First()
			if name == "" {
				return errors.New("project name required: fugo init <name>")
			}

			dir := filepath.Clean(name)
			out.tracef("template=%s  dir=%s", template, dir)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create directory: %w", err)
			}

			mainFile := filepath.Join(dir, "main.go")
			if err := os.WriteFile(mainFile, []byte(scaffoldMain(template, name)), 0o644); err != nil {
				return fmt.Errorf("write main.go: %w", err)
			}
			out.successf("wrote %s %s", mainFile, out.paint(cDim, "("+template+" template)"))

			modInit := exec.CommandContext(ctx, "go", "mod", "init", name)
			modInit.Dir = dir
			if err := out.runStep("Initializing Go module", modInit); err != nil {
				return fmt.Errorf("go mod init: %w", err)
			}

			fugoDir := fugoSrc
			if fugoDir == "" {
				fugoDir = findFugoRepo()
			}
			if fugoDir != "" {
				if err := addReplaceDirective(dir, fugoDir); err != nil {
					return err
				}
			} else {
				out.warnf("local fugo checkout not found — using the published module (pass --fugo-src to override)")
			}

			tidy := exec.CommandContext(ctx, "go", "mod", "tidy")
			tidy.Dir = dir
			if err := out.runStep("Resolving dependencies", tidy); err != nil {
				return fmt.Errorf("go mod tidy: %w", err)
			}

			out.infof("")
			out.successf("created %s%c", dir, os.PathSeparator)
			out.infof("  next: %s", out.paint(cBold, "cd "+dir+" && fugo run"))

			return nil
		},
	}
}

// addReplaceDirective appends a `replace github.com/sazardev/fugo => <rel>` line
// to the new project's go.mod so it builds against the local fugo checkout.
func addReplaceDirective(projectDir, fugoDir string) error {
	absFugo, _ := filepath.Abs(fugoDir)
	absProject, _ := filepath.Abs(projectDir)

	rel, err := filepath.Rel(absProject, absFugo)
	if err != nil {
		rel = fugoDir
	}
	relPath := strings.ReplaceAll(rel, "\\", "/")

	out.tracef("local fugo: %s  (replace => %s)", absFugo, relPath)

	goModFile := filepath.Join(projectDir, "go.mod")
	data, err := os.ReadFile(goModFile)
	if err != nil {
		return fmt.Errorf("read go.mod: %w", err)
	}

	data = append(data, []byte(fmt.Sprintf("\nreplace github.com/sazardev/fugo => %s\n", relPath))...)
	if err := os.WriteFile(goModFile, data, 0o644); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}

	out.successf("linked local fugo %s", out.paint(cDim, "(replace => "+relPath+")"))

	return nil
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
		Usage: "Build and run the Fugo app in the current directory",
		Description: `Build the Go app in the current directory and launch it: the gRPC server
starts, the Flutter render client is spawned (and built once if missing), and
the window opens. Press Ctrl+C to stop.

With --watch the Flutter window stays open while the Go server rebuilds and
reconnects on every .go change (in-memory state resets across reloads).

Examples:
  fugo run
  fugo run --watch
  fugo run --addr 127.0.0.1:9600
  fugo run -V               # verbose: trace the build and stream the app's logs`,
		Flags: append([]cli.Flag{
			&cli.StringFlag{
				Name:        "addr",
				Value:       "127.0.0.1:9510",
				Destination: &addr,
				Usage:       "gRPC listen address (host:port for TCP, a path for a Unix socket)",
			},
			&cli.StringFlag{
				Name:        "flutter",
				Destination: &flutter,
				Usage:       "path to the Flutter render binary (auto-detected if empty)",
			},
			&cli.BoolFlag{
				Name:        "watch",
				Aliases:     []string{"w"},
				Destination: &watch,
				Usage:       "rebuild the Go server on .go changes; keep the window open",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, _ *cli.Command) error {
			setupUI()

			if !hasMainGo() {
				return errors.New("no main.go in the current directory — run 'fugo init <name>' first")
			}

			out.infof("%s %s", out.paint(cBold, "Fugo"), out.paint(cDim, "v"+version))

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
	build := exec.CommandContext(ctx, "go", subcmdBuild, "-o", appBinary(), ".")
	if err := out.runStep("Building app", build); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	out.infof("%s running %s", out.paint(cBlue, "▶"), out.paint(cDim, "(Ctrl+C to stop)"))

	return runApp(ctx, addr, flutter)
}

func runWithWatch(ctx context.Context, addr, flutter string) error {
	out.infof("%s hot reload — window stays open; the Go server rebuilds on .go changes", out.paint(cCyan, "↻"))

	flutterProc, err := startFlutterClient(ctx, addr, flutter)
	if err != nil {
		out.warnf("could not start the Flutter client (%v) — falling back to full restarts", err)

		return runWithFullRestart(ctx, addr, flutter)
	}
	defer killProc(flutterProc)

	snap := fileSnapshot()
	for {
		if buildErr := buildApp(ctx); buildErr != nil {
			out.warnf("waiting for changes after build failure")
			waitForChange(&snap)

			continue
		}

		server := startServerOnly(ctx, addr)
		waitForChange(&snap)
		killProc(server)
		out.infof("%s change detected — reloading Go server", out.paint(cCyan, "↻"))
	}
}

func runWithFullRestart(ctx context.Context, addr, flutter string) error {
	out.infof("watching .go files for changes")

	snap := fileSnapshot()
	for {
		if err := buildAndRun(ctx, addr, flutter); err != nil {
			out.failf("%v", err)
		}

		waitForChange(&snap)
		out.infof("%s change detected — restarting", out.paint(cCyan, "↻"))
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

	return out.runStep("Building Go server", cmd)
}

// startServerOnly runs the built app in server-only mode (FUGO_NO_FLUTTER=1) so
// the externally-managed Flutter client reconnects to it across reloads.
func startServerOnly(ctx context.Context, addr string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, appBinary())
	cmd.Env = append(os.Environ(), "FUGO_ADDR="+addr, "FUGO_NO_FLUTTER=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		out.failf("start server: %v", err)

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
		Usage: "Build a release binary and bundle the Flutter client into dist/",
		Description: `Compile a stripped release binary and copy the precompiled Flutter render
client next to it under dist/. Ship the whole dist/ folder; it runs without a
Go or Flutter toolchain installed.

If the Flutter client bundle hasn't been built yet, only the Go binary is
produced (build the client once with: cd <fugo>/flutter_client && flutter build).

Examples:
  fugo build
  fugo build -V`,
		Flags: verbosityFlags(),
		Action: func(ctx context.Context, _ *cli.Command) error {
			setupUI()

			if !hasMainGo() {
				return errors.New("no main.go in the current directory — run 'fugo init <name>' first")
			}

			outDir := "dist"
			appOut := filepath.Join(outDir, projectName()+exeSuffix())

			build := exec.CommandContext(ctx, "go", subcmdBuild, "-ldflags=-s -w", "-o", appOut, ".")
			if err := out.runStep("Building app (release)", build); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}

			src := flutterBundleDir(ctx)
			if src == "" {
				out.warnf("Flutter client bundle not found — built the Go binary only")
				out.infof("  build the client once: cd <fugo>/flutter_client && flutter build %s", flutterTarget())
				out.successf("built %s %s", appOut, out.paint(cDim, "(app only)"))

				return nil
			}

			dst := filepath.Join(outDir, "flutter")
			start := time.Now()
			out.tracef("copy %s -> %s", src, dst)
			if err := copyDir(src, dst); err != nil {
				return fmt.Errorf("bundle flutter client: %w", err)
			}
			out.successf("bundled Flutter client %s", out.paint(cDim, "("+time.Since(start).Round(time.Millisecond).String()+")"))

			out.infof("")
			out.successf("build complete → %s%c", outDir, os.PathSeparator)
			out.infof("  %-9s your app", filepath.Base(appOut))
			out.infof("  %-9s bundled render client", "flutter"+string(os.PathSeparator))
			out.infof("  ship the whole %s%c folder; run: %s", outDir, os.PathSeparator, appOut)

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
		out.warnf("Flutter client not built and 'flutter' is not on PATH")
		out.infof("  build it once: cd flutter_client && flutter build %s", flutterTarget())

		return
	}

	args := []string{subcmdBuild, flutterTarget()}
	if runtime.GOOS != osWindows {
		args = append(args, "--debug")
	}

	cmd := exec.CommandContext(ctx, "flutter", args...)
	cmd.Dir = clientDir
	if err := out.runStep("Building Flutter client (first run — this can take a few minutes)", cmd); err != nil {
		out.failf("flutter build failed: %v", err)
	}
}

// flutterTarget is the `flutter build` target for the host OS.
func flutterTarget() string {
	if runtime.GOOS == osWindows {
		return "windows"
	}

	return "linux"
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
		Usage: "Check the development environment",
		Description: `Probe the toolchain Fugo needs and report what's found. Use -V for full
version output and the resolved fugo module path.`,
		Flags: verbosityFlags(),
		Action: func(ctx context.Context, _ *cli.Command) error {
			setupUI()

			checks := []struct {
				name string
				bin  string
				args []string
				hint string
			}{
				{"Go", "go", []string{"version"}, "required — https://go.dev/dl"},
				{"Flutter", "flutter", []string{"--version"}, "required to render — https://docs.flutter.dev"},
				{"protoc", "protoc", []string{"--version"}, "only to regenerate protobuf (make proto)"},
				{"gofumpt", "gofumpt", []string{"-version"}, "formatter — go install mvdan.cc/gofumpt@latest"},
			}

			out.heading("Fugo Doctor")

			missing := 0
			for _, c := range checks {
				line, err := firstLine(ctx, c.bin, c.args...)
				if err != nil {
					missing++
					out.printf("  %s %-9s %s\n", out.paint(cRed, "✗"), c.name, out.paint(cDim, c.hint))

					continue
				}

				out.printf("  %s %-9s %s\n", out.paint(cGreen, "✓"), c.name, line)
				out.tracef("%s: %s %s", c.name, c.bin, strings.Join(c.args, " "))
			}

			out.printf("\n  %-11s %s/%s\n", "platform", runtime.GOOS, runtime.GOARCH)
			if repo := fugoModuleDir(ctx); repo != "" {
				out.printf("  %-11s %s\n", "fugo module", repo)
			}

			out.printf("\n")
			if missing == 0 {
				out.successf("environment looks good")
			} else {
				out.warnf("%d tool(s) missing — see the hints above", missing)
			}

			return nil
		},
	}
}

// firstLine runs name with args and returns the trimmed first line of its
// combined output, or an error if the command cannot be run.
func firstLine(ctx context.Context, name string, args ...string) (string, error) {
	o, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
	if err != nil {
		return "", err
	}

	s := string(o)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}

	return strings.TrimSpace(s), nil
}
