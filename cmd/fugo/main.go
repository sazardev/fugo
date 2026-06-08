package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"
)

var (
	version = "0.1.0"
	commit  = "unknown"
	date    = "unknown"
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

func initCmd() *cli.Command {
	var fugoSrc string

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
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			name := c.Args().First()
			if name == "" {
				return fmt.Errorf("project name required: fugo init <name>")
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
			content := fmt.Sprintf(`package main

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
	counterText := fg.Text("0").
		FontSize(48).
		Color(fg.Hex("#FFFFFF"))

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
`, name)

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
		Action: func(ctx context.Context, c *cli.Command) error {
			if !hasMainGo() {
				return fmt.Errorf("no main.go found in current directory. Run 'fugo init <name>' first")
			}

			fmt.Printf("=== Fugo v%s ===\n", version)

			if watch {
				return runWithWatch(ctx, addr, flutter)
			}

			return buildAndRun(ctx, addr, flutter)
		},
	}
}

func buildAndRun(ctx context.Context, addr, flutter string) error {
	fmt.Println("Building...")
	build := exec.CommandContext(ctx, "go", "build", "-o", appBinary(), ".")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Println("Running... (press Ctrl+C to stop)")

	return runApp(ctx, addr, flutter)
}

func runWithWatch(ctx context.Context, addr, flutter string) error {
	fmt.Println("Watching .go files for changes...")

	snap := fileSnapshot()

	for {
		if err := buildAndRun(ctx, addr, flutter); err != nil {
			fmt.Printf("Run error: %v\n", err)
		}

		time.Sleep(500 * time.Millisecond)

		if changed := fileSnapshot(); !snapshotEq(snap, changed) {
			snap = changed
			fmt.Println("--- File change detected, restarting ---")
		} else {
			fmt.Println("App exited. Waiting for file changes...")
			for {
				time.Sleep(500 * time.Millisecond)
				changed = fileSnapshot()
				if !snapshotEq(snap, changed) {
					snap = changed
					fmt.Println("--- File change detected, restarting ---")

					break
				}
			}
		}
	}
}

func fileSnapshot() map[string]time.Time {
	snap := make(map[string]time.Time)

	_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
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
	run.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

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
		Action: func(ctx context.Context, c *cli.Command) error {
			if !hasMainGo() {
				return fmt.Errorf("no main.go found in current directory. Run 'fugo init <name>' first")
			}

			outDir := "dist"
			appOut := filepath.Join(outDir, projectName()+exeSuffix())

			fmt.Println("Building app (release)...")
			build := exec.CommandContext(ctx, "go", "build", "-ldflags=-s -w", "-o", appOut, ".")
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
	if runtime.GOOS == "windows" {
		return ".exe"
	}

	return ""
}

// flutterBundleDir locates the precompiled Flutter client bundle inside the
// fugo module — resolved via `go list -m` so it honors a replace directive —
// or "" if fugo is not a local checkout or the client has not been built yet.
func flutterBundleDir(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "go", "list", "-m", "-f", "{{.Dir}}", "github.com/sazardev/fugo").Output()
	if err != nil {
		return ""
	}

	repo := strings.TrimSpace(string(out))
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
	if runtime.GOOS == "windows" {
		return "bin/app.exe"
	}

	return "bin/app"
}

func doctorCmd() *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "Check development environment",
		Action: func(ctx context.Context, c *cli.Command) error {
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
				cmd := exec.Command(check.cmd, check.args...)
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
