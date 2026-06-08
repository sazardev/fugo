package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/urfave/cli/v3"
	xterm "golang.org/x/term"
)

// Verbosity flag destinations, shared across every subcommand (see
// verbosityFlags). They are read by setupUI at the start of each Action.
var (
	flagVerbose bool
	flagQuiet   bool
	flagNoColor bool
)

// verbosityFlags returns the logging flags attached to every subcommand so that
// `fugo <cmd> --verbose|-V` and `fugo <cmd> --quiet|-q` work for the command
// being run. Only one subcommand runs per invocation, so sharing the
// destinations is safe.
func verbosityFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Aliases:     []string{"V"},
			Destination: &flagVerbose,
			Usage:       "trace every step: commands run, resolved paths, timings, and the app's [fugo] logs",
		},
		&cli.BoolFlag{
			Name:        "quiet",
			Aliases:     []string{"q", "silent"},
			Destination: &flagQuiet,
			Usage:       "print only warnings and errors (silences progress and the app's logs)",
		},
		&cli.BoolFlag{
			Name:        "no-color",
			Destination: &flagNoColor,
			Usage:       "disable ANSI colors (also honored via NO_COLOR)",
		},
	}
}

type verbosity int

const (
	lvlQuiet   verbosity = iota // warnings + errors only
	lvlNormal                   // + progress / steps / success
	lvlVerbose                  // + traces and live subprocess output
)

// term is the CLI's terminal writer: leveled, optionally colored, with a step
// spinner. All status goes to stderr; machine-readable results go to stdout via
// printf so they can be piped cleanly.
type term struct {
	mu     sync.Mutex
	lvl    verbosity
	color  bool
	stderr io.Writer
	stdout io.Writer
}

var out = &term{lvl: lvlNormal, color: false, stderr: os.Stderr, stdout: os.Stdout}

// setupUI configures the global terminal from the parsed verbosity flags and
// exports FUGO_LOG so a spawned app honors the same verbosity. Call it first in
// every command Action.
func setupUI() {
	switch {
	case flagQuiet:
		out.lvl = lvlQuiet
	case flagVerbose:
		out.lvl = lvlVerbose
	default:
		out.lvl = lvlNormal
	}

	out.color = colorEnabled()

	// Propagate verbosity to the child app's [fugo] logger (see flog).
	switch out.lvl {
	case lvlQuiet:
		_ = os.Setenv("FUGO_LOG", "silent")
	case lvlVerbose:
		_ = os.Setenv("FUGO_LOG", "debug")
	case lvlNormal:
		_ = os.Setenv("FUGO_LOG", "info")
	}
}

func colorEnabled() bool {
	if flagNoColor || os.Getenv("NO_COLOR") != "" {
		return false
	}

	if !isTerminal(os.Stderr) {
		return false
	}

	enableVirtualTerminal(os.Stderr) // best-effort on Windows; no-op elsewhere

	return true
}

func isTerminal(f *os.File) bool {
	return xterm.IsTerminal(int(f.Fd()))
}

// ANSI color codes.
const (
	cReset  = "\x1b[0m"
	cBold   = "\x1b[1m"
	cDim    = "\x1b[2m"
	cRed    = "\x1b[31m"
	cGreen  = "\x1b[32m"
	cYellow = "\x1b[33m"
	cBlue   = "\x1b[34m"
	cCyan   = "\x1b[36m"
)

func (t *term) paint(code, s string) string {
	if !t.color {
		return s
	}

	return code + s + cReset
}

func (t *term) line(w io.Writer, s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, _ = fmt.Fprintln(w, s)
}

// tracef logs a dimmed diagnostic, shown only in verbose mode.
func (t *term) tracef(format string, a ...any) {
	if t.lvl < lvlVerbose {
		return
	}

	t.line(t.stderr, t.paint(cDim, "  · "+fmt.Sprintf(format, a...)))
}

// infof logs a normal-level status line (suppressed when quiet).
func (t *term) infof(format string, a ...any) {
	if t.lvl < lvlNormal {
		return
	}

	t.line(t.stderr, fmt.Sprintf(format, a...))
}

// successf logs a normal-level "✓" line (suppressed when quiet).
func (t *term) successf(format string, a ...any) {
	if t.lvl < lvlNormal {
		return
	}

	t.line(t.stderr, t.paint(cGreen, "✓ ")+fmt.Sprintf(format, a...))
}

// warnf logs a "!" line; shown at every level.
func (t *term) warnf(format string, a ...any) {
	t.line(t.stderr, t.paint(cYellow, "! ")+fmt.Sprintf(format, a...))
}

// failf logs a "✗" line; shown at every level.
func (t *term) failf(format string, a ...any) {
	t.line(t.stderr, t.paint(cRed, "✗ ")+fmt.Sprintf(format, a...))
}

// printf writes program output (results) to stdout, unconditionally.
func (t *term) printf(format string, a ...any) {
	_, _ = fmt.Fprintf(t.stdout, format, a...)
}

// heading prints a bold section heading to stdout.
func (t *term) heading(s string) {
	t.printf("%s\n", t.paint(cBold, s))
}

// runStep runs an external command as a labeled step. In verbose mode it traces
// the command and streams its output live; otherwise it shows an animated
// spinner, captures the output, and surfaces it only if the command fails. It
// always prints a ✓/✗ line with the elapsed time.
func (t *term) runStep(msg string, c *exec.Cmd) error {
	t.tracef("exec: %s%s", strings.Join(c.Args, " "), dirSuffix(c))
	start := time.Now()

	if t.lvl >= lvlVerbose {
		t.infof("%s %s", t.paint(cBlue, "▶"), msg)
		c.Stdout, c.Stderr = t.stderr, t.stderr
		err := c.Run()
		t.finishStep(msg, start, err)

		return err
	}

	var buf bytes.Buffer
	c.Stdout, c.Stderr = &buf, &buf
	stop := t.startSpin(msg, start)
	err := c.Run()
	stop()
	t.finishStep(msg, start, err)

	if err != nil && buf.Len() > 0 {
		t.mu.Lock()
		_, _ = fmt.Fprint(t.stderr, buf.String())
		if !strings.HasSuffix(buf.String(), "\n") {
			_, _ = fmt.Fprintln(t.stderr)
		}
		t.mu.Unlock()
	}

	return err
}

func (t *term) finishStep(msg string, start time.Time, err error) {
	el := time.Since(start).Round(time.Millisecond)
	if err != nil {
		t.failf("%s %s", msg, t.paint(cDim, "("+el.String()+")"))

		return
	}

	if t.lvl >= lvlNormal {
		t.line(t.stderr, fmt.Sprintf("%s %s %s", t.paint(cGreen, "✓"), msg, t.paint(cDim, "("+el.String()+")")))
	}
}

// startSpin animates a braille spinner with elapsed time until the returned
// stop function is called. It degrades to a single "▶" line when output is not
// an interactive terminal (e.g. CI) and to nothing when quiet.
func (t *term) startSpin(msg string, start time.Time) func() {
	if t.lvl < lvlNormal || !t.color || !isTerminal(os.Stderr) {
		if t.lvl >= lvlNormal {
			t.infof("%s %s", t.paint(cBlue, "▶"), msg)
		}

		return func() {}
	}

	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)
		frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		tick := time.NewTicker(90 * time.Millisecond)
		defer tick.Stop()

		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			case <-tick.C:
				el := time.Since(start).Round(time.Second)
				frame := string(frames[i%len(frames)])
				t.mu.Lock()
				_, _ = fmt.Fprintf(t.stderr, "\r%s %s %s", t.paint(cCyan, frame), msg, t.paint(cDim, "("+el.String()+")"))
				t.mu.Unlock()
			}
		}
	}()

	return func() {
		close(stop)
		<-done
		t.mu.Lock()
		_, _ = fmt.Fprint(t.stderr, "\r\x1b[K") // carriage return + clear to end of line
		t.mu.Unlock()
	}
}

func dirSuffix(c *exec.Cmd) string {
	if c.Dir != "" {
		return "  (dir=" + c.Dir + ")"
	}

	return ""
}
