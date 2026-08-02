// Command fugo is the Fugo CLI: init, run, build and doctor.
package main

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli/v3"

	"github.com/sazardev/fugo/config"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
)

const (
	osWindows   = "windows"
	subcmdBuild = "build"
	fugoModule  = "github.com/sazardev/fugo"
	versionFlag = "--version"
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
  fugo generate          scaffold a new screen or reusable component
  fugo vet               run Fugo's opinionated static analyzer
  fugo fix               auto-fix, format, and reorder by Fugo convention
  fugo doctor            check the toolchain + (in a project) its health
  fugo upgrade           update the fugo CLI to the latest release

Every command accepts -V/--verbose (trace commands, paths, timings and the
app's runtime logs) and -q/--quiet (errors only). Colors honor NO_COLOR.`,
		Commands: []*cli.Command{
			initCmd(),
			runCmd(),
			buildCmd(),
			widgetsCmd(),
			generateCmd(),
			vetCmd(),
			fixCmd(),
			doctorCmd(),
			autostartCmd(),
			upgradeCmd(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		out.failf("%v", err)
		os.Exit(1)
	}
}

// resolvedVersion is the plain semver fugo resolves itself to be — the same
// fallback chain versionString uses, without the "(commit ..., built ...)"
// suffix, so it can also drive the precompiled Flutter client download URL.
func resolvedVersion() string {
	v := version

	info, ok := debug.ReadBuildInfo()
	if ok && v == "0.1.0" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = strings.TrimPrefix(info.Main.Version, "v")
	}

	return v
}

// versionString reports the CLI version. `make build` injects version, commit
// and date through -ldflags; a `go install` binary keeps the defaults, so we
// fall back to the module version and VCS stamps the Go toolchain embeds in the
// build info, keeping `fugo --version` accurate either way.
func versionString() string {
	v, c, d := resolvedVersion(), commit, date

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("%s (commit %s, built %s)", v, c, d)
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

//nolint:gocognit // the wizard flow (flags vs. interactive prompts vs. defaults, each field independently optional) is inherently a lot of branches; splitting it up would just move them into more functions
func initCmd() *cli.Command {
	var (
		fugoSrc      string
		template     string
		theme        string
		organization string
		noGit        bool
		yes          bool
	)

	return &cli.Command{
		Name:      "init",
		Usage:     "Create a new Fugo project",
		ArgsUsage: "[project-name]",
		Description: `Scaffold a new Fugo project with a recommended layout: a thin main.go, a
ui package for your screens, fugo.toml for the window/server config, a README
and .gitignore, plus bin/ dist/ logs/ folders. It runs 'go mod init' + 'go mod
tidy' and initializes a git repo with an initial commit.

Run with no arguments in a terminal for the interactive wizard: it asks for
the project name, organization, template, and theme, then scaffolds
everything pre-configured and ready to run. Pass flags (or --yes) to skip
straight to scaffolding — the same path scripts and CI use.

Templates (--template, -t):
  counter   minimal counter — one screen, two FABs, live state (default)
  app       themed multi-page starter with a Router and Home/About pages
  showcase  most widgets on one scrollable page — a living API reference

Examples:
  fugo init                       interactive wizard
  fugo init myapp                 wizard still asks org/template/theme
  fugo init myapp -t showcase --theme dark --org com.acme -y
  fugo init myapp --no-git
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
			&cli.StringFlag{
				Name:        "theme",
				Destination: &theme,
				Usage:       "color scheme: light | dark (default: the template's own choice)",
			},
			&cli.StringFlag{
				Name:        "organization",
				Aliases:     []string{"org"},
				Destination: &organization,
				Usage:       "reverse-DNS organization (e.g. com.acme) — stored in fugo.toml, reserved for future packaging",
			},
			&cli.BoolFlag{
				Name:        "no-git",
				Destination: &noGit,
				Usage:       "skip 'git init' and the initial commit",
			},
			&cli.BoolFlag{
				Name:        "yes",
				Aliases:     []string{"y"},
				Destination: &yes,
				Usage:       "skip the interactive wizard; use flags/defaults for anything not passed",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			setupUI()

			name := c.Args().First()
			interactive := !yes && isTerminal(os.Stdin)

			if name == "" {
				if !interactive {
					return errors.New("project name required: fugo init <name> (or run with no arguments in a terminal for the wizard)")
				}

				out.heading("Fugo — new project")
			}

			w := newWizard()

			if name == "" {
				name = w.askRequired("Project name")
			}
			if interactive && !c.IsSet("organization") {
				organization = w.ask("Organization (reverse-DNS, e.g. com.acme, optional)", "")
			}
			if interactive && !c.IsSet("template") {
				template = w.askChoice("Template", []string{"counter", "app", "showcase"}, template)
			}
			if interactive && !c.IsSet("theme") {
				theme = w.askChoice("Theme", []string{"light", "dark"}, "")
			}
			if interactive && !c.IsSet("no-git") {
				noGit = !w.askYesNo("Initialize git repo?", true)
			}

			dir := filepath.Clean(name)
			module := filepath.Base(dir)
			files := filesFor(template, module)
			if theme != "" {
				files.theme = strings.ToUpper(theme[:1]) + theme[1:]
			}
			out.tracef("template=%s theme=%s org=%q dir=%s module=%s", template, files.theme, organization, dir, module)

			if err := scaffoldProject(dir, module, organization, files); err != nil {
				return err
			}
			out.successf("scaffolded %s %s", dir+string(os.PathSeparator), out.paint(cDim, "("+template+" template)"))

			modInit := exec.CommandContext(ctx, "go", "mod", "init", module)
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

			if !noGit {
				initGitRepo(ctx, dir)
			}

			printInitSummary(dir)

			return nil
		},
	}
}

// scaffoldProject writes the project's directory skeleton and source files.
func scaffoldProject(dir, module, organization string, files projectFiles) error {
	for _, d := range []string{dir, filepath.Join(dir, "ui"), filepath.Join(dir, "logs")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}

	writes := []struct{ path, content string }{
		{filepath.Join(dir, "main.go"), mainGo(module, files.theme)},
		{filepath.Join(dir, "ui", "home.go"), files.uiHome},
		{filepath.Join(dir, "fugo.toml"), fmt.Sprintf(configTemplate, module, module, files.width, files.height, appConfigBlock(organization))},
		{filepath.Join(dir, "README.md"), fmt.Sprintf(readmeTemplate, module)},
		{filepath.Join(dir, ".gitignore"), gitignoreTemplate},
		{filepath.Join(dir, "logs", ".gitkeep"), ""},
	}
	for _, w := range writes {
		if err := os.WriteFile(w.path, []byte(w.content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", w.path, err)
		}
		out.tracef("wrote %s", w.path)
	}

	return nil
}

// initGitRepo runs 'git init' + an initial commit in dir. A missing git binary
// or an unset user identity is non-fatal: the repo is left in place and the
// user is told what to do.
func initGitRepo(ctx context.Context, dir string) {
	if _, err := exec.LookPath("git"); err != nil {
		out.tracef("git not on PATH — skipping repo init")

		return
	}

	git := func(args ...string) error {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir

		return cmd.Run()
	}

	if err := git("init", "-q"); err != nil {
		out.warnf("git init failed: %v", err)

		return
	}
	_ = git("add", "-A")
	if err := git("commit", "-q", "-m", "chore: scaffold with fugo"); err != nil {
		out.warnf("git repo ready; initial commit skipped (set git user.name/user.email, then commit)")

		return
	}

	out.successf("initialized git repo %s", out.paint(cDim, "(initial commit)"))
}

// printInitSummary prints the generated layout and the next step.
func printInitSummary(dir string) {
	out.infof("")
	out.successf("created %s%c", dir, os.PathSeparator)
	for _, line := range []string{
		"main.go      entrypoint (theme + ui.Build)",
		"ui/home.go   your first screen",
		"fugo.toml    window + server config",
		"README.md    project readme",
		"logs/        runtime logs (gitignored)",
	} {
		out.infof("  %s", out.paint(cDim, line))
	}
	out.infof("")
	out.infof("  next: %s", out.paint(cBold, "cd "+dir+" && fugo run"))
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
		noWatch bool
		watch   bool // deprecated: hot reload is the default now; kept so --watch still parses
	)

	return &cli.Command{
		Name:  "run",
		Usage: "Build and run the Fugo app, hot-reloading on .go changes",
		Description: `Build the Go app in the current directory and launch it: the gRPC server
starts, the Flutter render client is spawned (and built once if missing), and
the window opens. Press Ctrl+C to stop.

Hot reload is ON by default — the Flutter window stays open while the Go server
rebuilds and reconnects on every .go change, so edits to text, handlers, layout,
etc. show up live (in-memory state resets across reloads). Pass --no-watch for a
single build-and-run.

Examples:
  fugo run                  # build, launch, and hot-reload on changes
  fugo run --no-watch       # build and run once (no auto-rebuild)
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
				Name:        "no-watch",
				Destination: &noWatch,
				Usage:       "disable hot reload; build and run once",
			},
			&cli.BoolFlag{
				Name:        "watch",
				Aliases:     []string{"w"},
				Destination: &watch,
				Hidden:      true,
				Usage:       "(deprecated) hot reload is the default; this flag is a no-op",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			setupUI()

			if !hasMainGo() {
				return errors.New("no main.go in the current directory — run 'fugo init <name>' first")
			}

			// Fall back to fugo.toml's [server] addr when --addr wasn't passed.
			if !c.IsSet("addr") {
				if a := config.Find(config.DefaultName).Server.Addr; a != "" {
					addr = a
				}
			}

			closeLog := setupRunLog()
			defer closeLog()

			out.infof("%s %s", out.paint(cBold, "Fugo"), out.paint(cDim, "v"+version))

			if flutter == "" {
				ensureFlutterClient(ctx)
			}

			if noWatch {
				return buildAndRun(ctx, addr, flutter)
			}

			return runWithWatch(ctx, addr, flutter)
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

	watcher, err := newWatcher()
	if err != nil {
		return fmt.Errorf("watch .go files: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	for {
		if buildErr := buildApp(ctx); buildErr != nil {
			out.warnf("fix the error above and save to retry")
			waitForChange(watcher)

			continue
		}

		server := startServerOnly(ctx, addr)
		waitForChange(watcher)
		killProc(server)
		out.infof("%s change detected — reloading Go server", out.paint(cCyan, "↻"))
	}
}

func runWithFullRestart(ctx context.Context, addr, flutter string) error {
	out.infof("watching .go files for changes")

	watcher, err := newWatcher()
	if err != nil {
		return fmt.Errorf("watch .go files: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	for {
		if err := buildAndRun(ctx, addr, flutter); err != nil {
			out.failf("%v", err)
		}

		waitForChange(watcher)
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
	cmd.Stdout = appLog
	cmd.Stderr = appLog

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
	cmd.Stdout = appLog
	cmd.Stderr = appLog
	if err := cmd.Start(); err != nil {
		out.failf("start server: %v", err)

		return nil
	}

	return cmd
}

// newWatcher builds an fsnotify watcher covering every directory in the
// project tree (fsnotify has no recursive mode), skipping build/VCS output so
// writes to bin/dist/logs don't trigger spurious reloads.
func newWatcher() (*fsnotify.Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	walkErr := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // intentional fallback: skip unreadable paths during walk
		}
		if !d.IsDir() {
			return nil
		}
		if isWatchExcluded(path) {
			return filepath.SkipDir
		}

		return w.Add(path)
	})
	if walkErr != nil {
		_ = w.Close()

		return nil, walkErr
	}

	return w, nil
}

func isWatchExcluded(path string) bool {
	switch filepath.Base(path) {
	case ".git", "bin", "dist", "logs", "vendor":
		return true
	default:
		return false
	}
}

// waitForChange blocks until a .go file is created, written, or removed.
// Saves often fire several fsnotify events in a burst (editors write via a
// temp file + rename, or gofmt rewrites right after a save), so events are
// debounced into a single wake-up instead of reloading once per event.
func waitForChange(w *fsnotify.Watcher) {
	const debounce = 150 * time.Millisecond

	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()

	pending := false

	for {
		select {
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if info, statErr := os.Stat(ev.Name); statErr == nil && info.IsDir() {
				if ev.Op&fsnotify.Create != 0 && !isWatchExcluded(ev.Name) {
					_ = w.Add(ev.Name)
				}

				continue
			}
			if filepath.Ext(ev.Name) != ".go" {
				continue
			}

			pending = true
			timer.Reset(debounce)
		case <-timer.C:
			if pending {
				return
			}
		case watchErr, ok := <-w.Errors:
			if !ok {
				return
			}

			out.warnf("watch error: %v", watchErr)
		}
	}
}

func killProc(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

// appLog is where a launched app's stdout/stderr go. setupRunLog points it at
// logs/run.log (tee'd to the console) for the duration of a `fugo run`.
var appLog io.Writer = os.Stdout

// setupRunLog tees the app's output to logs/run.log and returns a closer. If the
// file can't be created it falls back to console-only output.
func setupRunLog() func() {
	if err := os.MkdirAll("logs", 0o755); err != nil {
		out.tracef("logs: %v — console only", err)

		return func() {}
	}

	f, err := os.Create(filepath.Join("logs", "run.log"))
	if err != nil {
		out.tracef("logs: %v — console only", err)

		return func() {}
	}

	appLog = io.MultiWriter(os.Stdout, f)
	out.tracef("app output → logs%crun.log", os.PathSeparator)

	return func() {
		appLog = os.Stdout
		_ = f.Close()
	}
}

func runApp(ctx context.Context, addr, flutter string) error {
	run := exec.CommandContext(ctx, appBinary())
	run.Stdout = appLog
	run.Stderr = appLog
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

			// Ship fugo.toml beside the binary so the app keeps its window/server
			// config regardless of the launch directory.
			if _, err := os.Stat(config.DefaultName); err == nil {
				if err := copyFile(config.DefaultName, filepath.Join(outDir, config.DefaultName)); err != nil {
					out.tracef("copy %s: %v", config.DefaultName, err)
				}
			}

			if runtime.GOOS != osWindows {
				if err := writeLinuxPackaging(outDir, projectName(), appOut); err != nil {
					out.tracef("linux packaging: %v", err)
				}
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
			if runtime.GOOS != osWindows {
				out.infof("  %-9s XDG desktop entry (edit Icon= to point at your own icon)", projectName()+".desktop")
				out.infof("  %-9s installs the app + desktop entry to ~/.local", "install.sh")
				out.infof("  %-9s AUR packaging template — see its header comment", "PKGBUILD")
			}
			out.infof("  ship the whole %s%c folder; run: %s", outDir, os.PathSeparator, appOut)

			return nil
		},
	}
}

// writeLinuxPackaging writes a minimal set of Linux distribution artifacts
// next to appOut inside outDir: an XDG .desktop entry, a user-local
// install.sh (the realistic path for "install this app" without root or a
// distro package manager), and an AUR PKGBUILD template. None of this
// requires appimagetool/linuxdeploy or any other external tool — it's plain
// text generated from what 'fugo build' already knows.
func writeLinuxPackaging(outDir, name, appOut string) error {
	binName := filepath.Base(appOut)

	desktop := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s
Icon=utilities-terminal
Terminal=false
Categories=Utility;
`, name, binName)
	if err := os.WriteFile(filepath.Join(outDir, name+".desktop"), []byte(desktop), 0o644); err != nil {
		return err
	}

	install := fmt.Sprintf(`#!/bin/sh
# Installs %[1]s for the current user only (no root needed): the binary and
# bundled Flutter client go under ~/.local/share/%[1]s, and a desktop entry
# under ~/.local/share/applications so it shows up in app launchers.
set -e
dest="$HOME/.local/share/%[1]s"
mkdir -p "$dest" "$HOME/.local/share/applications" "$HOME/.local/bin"
cp -r "$(dirname "$0")"/* "$dest/"
ln -sf "$dest/%[2]s" "$HOME/.local/bin/%[1]s"
sed "s#Exec=%[2]s#Exec=$dest/%[2]s#" "$dest/%[1]s.desktop" > "$HOME/.local/share/applications/%[1]s.desktop"
echo "Installed. Run '%[1]s' (make sure ~/.local/bin is on PATH) or launch it from your app menu."
`, name, binName)
	if err := os.WriteFile(filepath.Join(outDir, "install.sh"), []byte(install), 0o755); err != nil {
		return err
	}

	pkgbuild := fmt.Sprintf(`# Maintainer: you <you@example.com>
# AUR packaging template for %[1]s — a Fugo app. Fill in pkgver/source/sha256sums
# for a real release tarball (this points at the local dist/ build as a
# starting point) and drop this at the root of your AUR git repo.
pkgname=%[1]s
pkgver=1.0.0
pkgrel=1
pkgdesc="%[1]s"
arch=('x86_64')
url="https://example.com/%[1]s"
license=('unknown')
depends=('gtk3')
source=("%[1]s-$pkgver.tar.gz::file://%[1]s")
sha256sums=('SKIP')

package() {
  install -Dm755 "$srcdir/%[1]s/%[2]s" "$pkgdir/usr/bin/%[1]s"
  install -Dm644 "$srcdir/%[1]s/%[1]s.desktop" "$pkgdir/usr/share/applications/%[1]s.desktop"
  cp -r "$srcdir/%[1]s/flutter" "$pkgdir/usr/lib/%[1]s" 2>/dev/null || true
}
`, name, binName)

	return os.WriteFile(filepath.Join(outDir, "PKGBUILD"), []byte(pkgbuild), 0o644)
}

// projectName returns the app/binary name: fugo.toml's name when set to a real
// value, otherwise the current directory's base name.
func projectName() string {
	if cfg := config.Find(config.DefaultName); cfg.Name != "" && cfg.Name != config.Default().Name {
		return cfg.Name
	}

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

// flutterBundleDir locates a usable precompiled Flutter client bundle: first
// a cached download for the running CLI's version (see flutterdl.go — no
// Flutter SDK involved), then a local build inside the fugo module (resolved
// via fugoModuleDir so it honors a replace directive). "" if neither exists
// yet.
func flutterBundleDir(ctx context.Context) string {
	if dir := downloadedFlutterClientDir(resolvedVersion()); dir != "" {
		return dir
	}

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

// ensureFlutterClient gets a Flutter render client ready without the user
// installing anything extra: it first tries downloading the precompiled
// client that matches this CLI's own release (see flutterdl.go), and only
// falls back to a local 'flutter build' — which does require the Flutter SDK
// — when no matching download exists (an unreleased/dev build of fugo, no
// network, or a platform without a published binary) or the fugo source tree
// can't be located. Flutter SDK stays useful for anyone adding custom Flutter
// packages to the client and rebuilding it themselves; it's just no longer
// required for the default path. No-op if a bundle is already available.
func ensureFlutterClient(ctx context.Context) {
	if flutterBundleDir(ctx) != "" {
		return
	}

	// "0.1.0" is the unresolved placeholder in `var version` (see the top of
	// this file) — a plain `go build` of this repo without make's -ldflags,
	// not a real tagged release. Skip straight to the local-build fallback
	// instead of trying (and failing) a real network request against a
	// release that doesn't exist.
	if v := resolvedVersion(); v != "0.1.0" {
		dir, err := downloadFlutterClient(ctx, v)
		if err == nil {
			out.successf("downloaded the precompiled Flutter client %s", out.paint(cDim, "("+dir+")"))

			return
		}
		out.tracef("precompiled Flutter client not used: %v", err)
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
		Usage: "Check the toolchain and (inside a project) the project's health",
		Description: `Diagnose what Fugo needs to build and run. It checks the toolchain (Go,
Flutter, git, protoc, gofumpt) and, when run inside a Fugo project, imports
fugo.toml and validates the project: structure (main.go, ui/, go.mod), that the
fugo module resolves, whether the Flutter client is built, whether the
configured address is free, and that the project compiles.

Exits non-zero if there is a blocking issue (✗), so it works in scripts/CI.
Use -V to trace each probe.

Also runs Fugo's opinionated analyzer (fugovet) as a non-blocking check — a
convention finding is a warning, not a build error. 'fugo vet' shows details,
'fugo fix' auto-fixes what it safely can.

With --fix it first repairs the auto-fixable bits inside a project: writes a
default fugo.toml if missing, runs 'git init' if there's no repo, 'go mod
tidy' to resolve dependencies, applies fugovet's mechanical fixes, and runs
gofumpt — then re-runs the diagnosis. (It does not reorder declarations;
that's 'fugo fix''s opt-in structural pass.)`,
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:  "fix",
				Usage: "auto-repair what it can (fugo.toml, git init, go mod tidy) before checking",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			setupUI()
			out.heading("Fugo Doctor")

			if c.Bool("fix") && inProject() {
				doctorFix(ctx)
			}

			rep := &doctorReport{}
			doctorToolchain(ctx, rep)

			if inProject() {
				doctorProject(ctx, rep)
			} else {
				out.printf("\n  %s %s\n", out.paint(cDim, "·"),
					out.paint(cDim, "not inside a Fugo project — run 'fugo init <name>' to scaffold one"))
			}

			out.printf("\n")
			switch {
			case rep.fails > 0:
				out.failf("%d blocking issue(s), %d warning(s) — fix the ✗ items above", rep.fails, rep.warns)

				return fmt.Errorf("doctor: %d blocking issue(s)", rep.fails)
			case rep.warns > 0:
				out.warnf("%d warning(s) — see the ! items above", rep.warns)
			default:
				out.successf("all good — you're ready to 'fugo run'")
			}

			return nil
		},
	}
}

// doctorReport tallies ✗/! findings and prints aligned status lines.
type doctorReport struct {
	fails int
	warns int
}

func (r *doctorReport) ok(label, detail string) {
	out.printf("  %s %-14s %s\n", out.paint(cGreen, "✓"), label, detail)
}

func (r *doctorReport) warn(label, hint string) {
	r.warns++
	out.printf("  %s %-14s %s\n", out.paint(cYellow, "!"), label, out.paint(cDim, hint))
}

func (r *doctorReport) fail(label, hint string) {
	r.fails++
	out.printf("  %s %-14s %s\n", out.paint(cRed, "✗"), label, out.paint(cDim, hint))
}

func (r *doctorReport) note(label, detail string) {
	out.printf("  %s %-14s %s\n", out.paint(cDim, "·"), label, out.paint(cDim, detail))
}

// doctorToolchain checks the external tools Fugo relies on.
func doctorToolchain(ctx context.Context, rep *doctorReport) {
	out.printf("\n%s\n", out.paint(cBold, "Toolchain"))

	tools := []struct {
		name, bin string
		args      []string
		required  bool
		hint      string
	}{
		{"Go", "go", []string{"version"}, true, "required — https://go.dev/dl"},
		{"Flutter", "flutter", []string{versionFlag}, false, "optional — 'fugo run' downloads a precompiled client automatically; install to add custom Flutter packages or build for an unpublished platform"},
		{"git", "git", []string{versionFlag}, false, "recommended — 'fugo init' starts a repo"},
		{"protoc", "protoc", []string{versionFlag}, false, "only to regenerate protobuf (make proto)"},
		{"gofumpt", "gofumpt", []string{"-version"}, false, "formatter — go install mvdan.cc/gofumpt@latest"},
	}
	for _, c := range tools {
		line, err := firstLine(ctx, c.bin, c.args...)
		switch {
		case err == nil:
			rep.ok(c.name, line)
		case c.required:
			rep.fail(c.name, c.hint)
		default:
			rep.warn(c.name, c.hint)
		}
	}

	doctorFlutterVersion(ctx, rep)
	doctorNotifications(ctx, rep)

	rep.note("platform", runtime.GOOS+"/"+runtime.GOARCH)
}

// doctorNotifications warns (never fails) when Context.Notifications().Show
// is unlikely to work: on Linux it's backed by libnotify over D-Bus, and
// there's no reliable way to query the D-Bus session bus for a running
// org.freedesktop.Notifications service without a D-Bus client library, so
// this checks for notify-send (part of the libnotify-bin/libnotify-tools
// package on most distros) as a proxy for "libnotify is installed".
func doctorNotifications(ctx context.Context, rep *doctorReport) {
	if runtime.GOOS != "linux" {
		return
	}

	if _, err := firstLine(ctx, "notify-send", versionFlag); err != nil {
		rep.warn("notifications", "notify-send not found — Context.Notifications().Show needs libnotify (e.g. 'libnotify-bin' on Debian/Ubuntu, 'libnotify' on Arch) and a running notification daemon (most desktop environments ship one)")

		return
	}

	rep.ok("notifications", "libnotify found")
}

// doctorFlutterVersion compares an installed Flutter SDK against
// FLUTTER_VERSION — the version fugo's precompiled client and flutter_client/
// are built against (see CLAUDE.md's versioning section). A mismatch is a
// warning, not a failure: 'fugo run' still works via the precompiled
// download; it only matters if you rebuild flutter_client/ yourself.
func doctorFlutterVersion(ctx context.Context, rep *doctorReport) {
	line, err := firstLine(ctx, "flutter", versionFlag)
	if err != nil {
		return // already reported (fail or warn) by the toolchain loop above
	}

	installed := parseFlutterVersion(line)
	if installed == "" {
		return
	}

	pinned := pinnedFlutterVersion(ctx)
	if pinned == "" {
		return // fugo module source not resolved — nothing to compare against
	}

	if installed == pinned {
		rep.ok("flutter version", installed+" (matches FLUTTER_VERSION)")

		return
	}

	rep.warn("flutter version", fmt.Sprintf(
		"%s installed, fugo targets %s — only matters if you rebuild flutter_client/ yourself ('fugo run' otherwise uses the precompiled client)",
		installed, pinned,
	))
}

// parseFlutterVersion extracts "3.44.8" from `flutter --version`'s first
// line ("Flutter 3.44.8 • channel stable • ...").
func parseFlutterVersion(line string) string {
	fields := strings.Fields(line)
	for i, f := range fields {
		if f == "Flutter" && i+1 < len(fields) {
			return fields[i+1]
		}
	}

	return ""
}

// pinnedFlutterVersion reads FLUTTER_VERSION from the resolved fugo module
// source (honoring a replace directive, same as fugoModuleDir), "" if it
// can't be resolved or read.
func pinnedFlutterVersion(ctx context.Context) string {
	repo := fugoModuleDir(ctx)
	if repo == "" {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(repo, "FLUTTER_VERSION"))
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// inProject reports whether the working directory looks like a Fugo project.
func inProject() bool {
	for _, f := range []string{config.DefaultName, "main.go"} {
		if _, err := os.Stat(f); err == nil {
			return true
		}
	}

	return false
}

// doctorProject imports fugo.toml and validates the project that would run.
func doctorProject(ctx context.Context, rep *doctorReport) {
	out.printf("\n%s\n", out.paint(cBold, "Project"))

	// fugo.toml — the project-level config the CLI imports.
	if _, err := os.Stat(config.DefaultName); err == nil {
		cfg, loadErr := config.Load(config.DefaultName)
		if loadErr != nil {
			rep.warn("fugo.toml", "unreadable, using defaults: "+loadErr.Error())
		} else {
			rep.ok("fugo.toml", fmt.Sprintf("%q  %dx%d  %s",
				cfg.Window.Title, cfg.Window.Width, cfg.Window.Height, cfg.Server.Addr))
			doctorAddr(ctx, cfg.Server.Addr, rep)
		}
	} else {
		rep.warn("fugo.toml", "missing — run/build fall back to defaults (800x600, "+config.DefaultAddr+")")
	}

	// Structure.
	doctorPath(rep, "main.go", true, "missing — not a Fugo project root?")
	doctorPath(rep, "go.mod", true, "missing — run 'go mod init <name>'")

	// The fugo module must resolve (honoring any replace directive).
	if repo := fugoModuleDir(ctx); repo != "" {
		rep.ok("fugo module", repo)
	} else {
		rep.fail("fugo module", "github.com/sazardev/fugo not resolved — run 'go mod tidy' (or 'fugo doctor --fix')")
	}

	// Coherence: go.mod module ↔ main.go's ui import ↔ ui.Build.
	doctorCoherence(rep)

	// Fugo's opinionated static analyzer — non-blocking, since it's a
	// convention check, not a build error.
	doctorFugovet(ctx, rep)

	// Flutter render client: ready (downloaded or locally built), or
	// obtainable on first run — via the precompiled download for a real
	// release, or a local 'flutter build' otherwise.
	_, lookErr := exec.LookPath("flutter")

	switch {
	case flutterBundleDir(ctx) != "":
		rep.ok("flutter client", "ready")
	case resolvedVersion() != "0.1.0":
		rep.note("flutter client", "not built yet — 'fugo run' downloads the precompiled client on first launch")
	case lookErr == nil:
		rep.note("flutter client", "not built yet — 'fugo run' builds it locally on first launch")
	default:
		rep.warn("flutter client", "not built, no precompiled client for a dev build of fugo, and flutter not on PATH — 'fugo run' cannot render")
	}

	// The definitive "will it run?" check.
	if err := out.runStep("Compiling project (go build ./...)", exec.CommandContext(ctx, "go", subcmdBuild, "./...")); err != nil {
		rep.fails++
	}
}

// doctorFugovet runs Fugo's opinionated static analyzer (fugovet) as a
// non-blocking check — a convention finding is a warning, not a build error,
// so it never fails 'fugo doctor' on its own.
func doctorFugovet(ctx context.Context, rep *doctorReport) {
	bin, err := ensureFugovet(ctx)
	if err != nil {
		rep.note("fugo vet", "skipped — "+err.Error())

		return
	}

	report, runErr := exec.CommandContext(ctx, bin, "./...").CombinedOutput()
	trimmed := strings.TrimSpace(string(report))

	switch {
	case runErr == nil:
		rep.ok("fugo vet", "no issues")
	case trimmed == "":
		rep.warn("fugo vet", "analyzer exited with an error and no output — run 'fugo vet' directly")
	default:
		n := len(strings.Split(trimmed, "\n"))
		rep.warn("fugo vet", fmt.Sprintf("%d finding(s) — 'fugo vet' for details, 'fugo fix' to auto-fix", n))
	}
}

// doctorPath checks for a project file/dir; missing required entries are ✗.
func doctorPath(rep *doctorReport, name string, required bool, hint string) {
	if _, err := os.Stat(name); err == nil {
		rep.ok(name, "present")

		return
	}
	if required {
		rep.fail(name, hint)
	} else {
		rep.warn(name, hint)
	}
}

// doctorAddr reports whether the configured gRPC address is usable: a free TCP
// port, or a Unix socket path. An in-use port is a warning (another instance).
func doctorAddr(ctx context.Context, addr string, rep *doctorReport) {
	if !strings.Contains(addr, ":") {
		rep.ok("server addr", addr+" (unix socket)")

		return
	}

	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		rep.warn("server addr", addr+" in use — stop the other instance or set a different [server] addr / --addr")

		return
	}
	_ = ln.Close()
	rep.ok("server addr", addr+" (free)")
}

// doctorCoherence checks that the scaffolded wiring still lines up: go.mod's
// module path, main.go's `<module>/ui` import, and a Build func in that package.
// It only reports when main.go actually imports a `/ui` subpackage, so flat or
// custom layouts are left alone.
func doctorCoherence(rep *doctorReport) {
	module := goModModule("go.mod")
	if module == "" {
		return // go.mod absence is already reported by the structure check.
	}

	uiImport := ""
	for _, imp := range goImports("main.go") {
		if strings.HasSuffix(imp, "/ui") {
			uiImport = imp

			break
		}
	}
	if uiImport == "" {
		return // not using a ui subpackage — nothing to reconcile.
	}

	if !strings.HasPrefix(uiImport, module+"/") {
		rep.fail("ui import", fmt.Sprintf("main.go imports %q but go.mod module is %q — make them match", uiImport, module))

		return
	}

	rel := strings.TrimPrefix(uiImport, module+"/")
	if _, err := os.Stat(rel); err != nil {
		rep.fail("ui package", fmt.Sprintf("main.go imports %s but ./%s is missing", uiImport, rel))

		return
	}
	if !hasBuildFunc(rel) {
		rep.fail("ui.Build", fmt.Sprintf("package %s has no exported Build(ctx) — main.go calls ui.Build", uiImport))

		return
	}

	rep.ok("ui import", uiImport+" → ./"+rel)
}

// goModModule returns the module path declared in a go.mod file ("" if absent).
func goModModule(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}

	return ""
}

// goImports returns the import paths of a single Go file ("nil" if it can't be
// parsed); only the import block is read.
func goImports(path string) []string {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		return nil
	}

	paths := make([]string, 0, len(f.Imports))
	for _, imp := range f.Imports {
		paths = append(paths, strings.Trim(imp.Path.Value, `"`))
	}

	return paths
}

// hasBuildFunc reports whether any non-test .go file in dir declares a
// receiver-less exported func named Build.
func hasBuildFunc(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, 0)
		if err != nil {
			continue
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "Build" {
				return true
			}
		}
	}

	return false
}

// doctorFix repairs the auto-fixable parts of a project before diagnosis: a
// missing fugo.toml, a missing git repo, and unresolved dependencies.
func doctorFix(ctx context.Context) {
	out.printf("\n%s\n", out.paint(cBold, "Fixing"))

	if _, err := os.Stat(config.DefaultName); err != nil {
		name := goModModule("go.mod")
		if name == "" {
			name = projectName()
		}
		content := fmt.Sprintf(configTemplate, filepath.Base(name), filepath.Base(name), 800, 600, "")
		if writeErr := os.WriteFile(config.DefaultName, []byte(content), 0o644); writeErr == nil {
			out.successf("wrote %s", config.DefaultName)
		} else {
			out.warnf("could not write %s: %v", config.DefaultName, writeErr)
		}
	}

	if _, err := os.Stat(".git"); err != nil {
		if _, lookErr := exec.LookPath("git"); lookErr == nil {
			_ = out.runStep("git init", exec.CommandContext(ctx, "git", "init", "-q"))
		}
	}

	if _, err := os.Stat("go.mod"); err == nil {
		_ = out.runStep("go mod tidy", exec.CommandContext(ctx, "go", "mod", "tidy"))

		if bin, err := ensureFugovet(ctx); err == nil {
			_ = out.runStep("Applying mechanical fixes (fugovet -fix)", exec.CommandContext(ctx, bin, "-fix", "./..."))
		}

		_ = gofumptProject(ctx)
	}
}

// upgradeCmd updates the fugo CLI itself to the latest published release.
func upgradeCmd() *cli.Command {
	return &cli.Command{
		Name:      "upgrade",
		Usage:     "Update the fugo CLI to the latest release",
		ArgsUsage: "[version]",
		Description: `Reinstall the fugo CLI with the Go toolchain, defaulting to the latest
release. Pass a version to pin one, e.g. fugo upgrade v0.4.2.

Requires Go on PATH; installs to $(go env GOBIN) or $(go env GOPATH)/bin. On
Windows the running binary is moved aside (<exe>.old) so it can be replaced.
This updates the CLI only — rebuild the Flutter client with 'flutter build'.`,
		Flags: verbosityFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			setupUI()

			return runUpgrade(ctx, cmd.Args().First())
		},
	}
}

func runUpgrade(ctx context.Context, version string) error {
	if _, err := exec.LookPath("go"); err != nil {
		out.failf("the Go toolchain is required to upgrade — install it from https://go.dev/dl")

		return errors.New("go toolchain not found on PATH")
	}

	version = strings.TrimPrefix(strings.TrimSpace(version), "@")
	if version == "" {
		version = "latest"
	}

	out.infof("current: fugo %s", versionString())

	// On Windows a running .exe cannot be overwritten, so move ourselves aside
	// first when go install would replace the binary we are running from.
	restore := stashRunningBinary(ctx)

	pkg := fugoModule + "/cmd/fugo@" + version
	if err := out.runStep("go install "+pkg, exec.CommandContext(ctx, "go", "install", pkg)); err != nil {
		restore()
		out.failf("upgrade failed — your existing fugo is unchanged")

		return err
	}

	out.successf("fugo upgraded (%s)", version)
	if dir := installDir(ctx); dir != "" {
		out.infof("installed to %s — run `fugo --version` to confirm (keep that dir on PATH)", dir)
	}

	return nil
}

// stashRunningBinary, on Windows, renames the currently-running fugo.exe to
// <exe>.old so `go install` can write a fresh one (Windows cannot overwrite a
// running image). It returns a function that restores the old binary, called if
// the install fails. On other systems it is a no-op — a running file can be
// replaced in place.
func stashRunningBinary(ctx context.Context) func() {
	noop := func() {}
	if runtime.GOOS != osWindows {
		return noop
	}

	self, err := os.Executable()
	if err != nil {
		return noop
	}

	if !samePath(self, filepath.Join(installDir(ctx), "fugo.exe")) {
		return noop // running from elsewhere (e.g. ./bin); go install won't touch us
	}

	bak := self + ".old"
	_ = os.Remove(bak) // clear a leftover from a previous upgrade
	if err := os.Rename(self, bak); err != nil {
		out.tracef("could not move the running binary aside: %v", err)

		return noop
	}

	out.tracef("moved running binary to %s", bak)

	return func() { _ = os.Rename(bak, self) }
}

// samePath reports whether two filesystem paths point to the same location,
// case-insensitively on Windows.
func samePath(a, b string) bool {
	ca, cb := filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == osWindows {
		return strings.EqualFold(ca, cb)
	}

	return ca == cb
}

// installDir reports where `go install` places binaries: $GOBIN, else
// $GOPATH/bin.
func installDir(ctx context.Context) string {
	if b := goEnv(ctx, "GOBIN"); b != "" {
		return b
	}

	if p := goEnv(ctx, "GOPATH"); p != "" {
		return filepath.Join(p, "bin")
	}

	return ""
}

// goEnv returns a single `go env` value, trimmed, or "" if it cannot be read.
func goEnv(ctx context.Context, key string) string {
	b, err := exec.CommandContext(ctx, "go", "env", key).Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(b))
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
