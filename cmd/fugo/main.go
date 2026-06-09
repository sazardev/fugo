// Command fugo is the Fugo CLI: init, run, build and doctor.
package main

import (
	"context"
	"errors"
	"fmt"
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
  fugo doctor            check the toolchain + (in a project) its health
  fugo upgrade           update the fugo CLI to the latest release

Every command accepts -V/--verbose (trace commands, paths, timings and the
app's runtime logs) and -q/--quiet (errors only). Colors honor NO_COLOR.`,
		Commands: []*cli.Command{
			initCmd(),
			runCmd(),
			buildCmd(),
			widgetsCmd(),
			doctorCmd(),
			upgradeCmd(),
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

func initCmd() *cli.Command {
	var (
		fugoSrc  string
		template string
		noGit    bool
	)

	return &cli.Command{
		Name:      "init",
		Usage:     "Create a new Fugo project",
		ArgsUsage: "<project-name>",
		Description: `Scaffold a new Fugo project with a recommended layout: a thin main.go, a
ui package for your screens, fugo.toml for the window/server config, a README
and .gitignore, plus bin/ dist/ logs/ folders. It runs 'go mod init' + 'go mod
tidy' and initializes a git repo with an initial commit.

Templates (--template, -t):
  counter   minimal counter — one screen, two FABs, live state (default)
  app       themed multi-page starter with a Router and Home/About pages
  showcase  most widgets on one scrollable page — a living API reference

Examples:
  fugo init myapp
  fugo init myapp -t showcase
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
			&cli.BoolFlag{
				Name:        "no-git",
				Destination: &noGit,
				Usage:       "skip 'git init' and the initial commit",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			setupUI()

			name := c.Args().First()
			if name == "" {
				return errors.New("project name required: fugo init <name>")
			}

			dir := filepath.Clean(name)
			module := filepath.Base(dir)
			files := filesFor(template, module)
			out.tracef("template=%s  dir=%s  module=%s", template, dir, module)

			if err := scaffoldProject(dir, module, files); err != nil {
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
func scaffoldProject(dir, module string, files projectFiles) error {
	for _, d := range []string{dir, filepath.Join(dir, "ui"), filepath.Join(dir, "logs")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}

	writes := []struct{ path, content string }{
		{filepath.Join(dir, "main.go"), mainGo(module, files.theme)},
		{filepath.Join(dir, "ui", "home.go"), files.uiHome},
		{filepath.Join(dir, "fugo.toml"), fmt.Sprintf(configTemplate, module, module, files.width, files.height)},
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
			if base == ".git" || base == "bin" || base == "dist" || base == "logs" || base == "vendor" {
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
		Usage: "Check the toolchain and (inside a project) the project's health",
		Description: `Diagnose what Fugo needs to build and run. It checks the toolchain (Go,
Flutter, git, protoc, gofumpt) and, when run inside a Fugo project, imports
fugo.toml and validates the project: structure (main.go, ui/, go.mod), that the
fugo module resolves, whether the Flutter client is built, whether the
configured address is free, and that the project compiles.

Exits non-zero if there is a blocking issue (✗), so it works in scripts/CI.
Use -V to trace each probe.`,
		Flags: verbosityFlags(),
		Action: func(ctx context.Context, _ *cli.Command) error {
			setupUI()
			out.heading("Fugo Doctor")

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
		{"Flutter", "flutter", []string{versionFlag}, true, "required to render — https://docs.flutter.dev"},
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

	rep.note("platform", runtime.GOOS+"/"+runtime.GOARCH)
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
	doctorPath(rep, "ui", false, "no ui/ package (older/flat layout) — fine if screens live elsewhere")
	doctorPath(rep, "go.mod", true, "missing — run 'go mod init <name>'")

	// The fugo module must resolve (honoring any replace directive).
	if repo := fugoModuleDir(ctx); repo != "" {
		rep.ok("fugo module", repo)
	} else {
		rep.fail("fugo module", "github.com/sazardev/fugo not resolved — run 'go mod tidy'")
	}

	// Flutter render client: built, or buildable on first run.
	if dir := flutterBundleDir(ctx); dir != "" {
		rep.ok("flutter client", "built")
	} else if _, err := exec.LookPath("flutter"); err == nil {
		rep.note("flutter client", "not built yet — 'fugo run' builds it on first launch")
	} else {
		rep.warn("flutter client", "not built and flutter not on PATH — 'fugo run' cannot render")
	}

	// The definitive "will it run?" check.
	if err := out.runStep("Compiling project (go build ./...)", exec.CommandContext(ctx, "go", subcmdBuild, "./...")); err != nil {
		rep.fails++
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
