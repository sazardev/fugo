package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"
)

const fgPkg = "github.com/sazardev/fugo/fg"

type widgetGroup struct {
	name  string
	items []widgetInfo
}

type widgetInfo struct {
	ctor string // prefix-free constructor, e.g. "Text" → fg.Text(...)
	desc string
}

// widgetCatalog is a curated, offline view of the fg widget API. `fugo widgets
// <name>` resolves the real doc comments via `go doc`; this list always works.
var widgetCatalog = []widgetGroup{
	{"Text & media", []widgetInfo{
		{"Text", "styled text run — FontSize, Color, Weight, Align"},
		{"Icon", "a Material icon by name — Size, Color"},
		{"Image", "a network image — Width, Height"},
		{"Divider", "a horizontal rule — Thickness, Color"},
	}},
	{"Layout", []widgetInfo{
		{"Container", "single-child box — BgColor, Pad, BorderRadius"},
		{"Row", "lay children out horizontally — MainAlign, CrossAlign"},
		{"Column", "lay children out vertically"},
		{"Stack", "overlap children; place them with Positioned"},
		{"Wrap", "flow children, wrapping to new lines — Spacing, RunSpacing"},
		{"Center", "center a single child"},
		{"Padding / PaddingAll", "inset a child by EdgeInsets"},
		{"SizedBox", "fixed-size box or a gap"},
		{"Expanded", "fill the remaining space in a Row/Column — Flex"},
		{"Positioned", "place a child inside a Stack — Left/Top/Right/Bottom"},
		{"Align", "align a child within itself"},
	}},
	{"Scrolling & lists", []widgetInfo{
		{"ListView", "scrollable list of children"},
		{"GridView", "grid of children — CrossAxisCount, ChildAspectRatio"},
		{"ScrollView", "make any child scrollable"},
	}},
	{"Input", []widgetInfo{
		{"Button", "clickable button — BgColor, OnClick"},
		{"TextField", "text input — placeholder, OnChange"},
		{"Checkbox", "boolean checkbox — OnChange"},
		{"Switch", "boolean toggle — OnChange"},
		{"Slider", "numeric slider — SetMin/SetMax/SetValue, OnChange"},
		{"Radio", "radio button in a Group — OnChange"},
		{"Dropdown", "select from a list — SetValue, OnChange"},
	}},
	{"Interaction & animation", []widgetInfo{
		{"GestureDetector", "wrap a child to capture taps — OnTap"},
		{"AnimatedContainer", "Container that animates property changes — DurationMs"},
		{"AnimatedOpacity", "fade a child in/out — DurationMs"},
	}},
	{"Routing", []widgetInfo{
		{"Router", "map routes to pages; navigate via ctx.NavigateTo / ctx.GoBack"},
	}},
}

func widgetsCmd() *cli.Command {
	var all bool

	return &cli.Command{
		Name:      "widgets",
		Usage:     "Browse the fg widget catalog and their doc comments",
		ArgsUsage: "[widget]",
		Description: `List the built-in fg widgets, or show one widget's own Go doc comment and its
chainable setters via 'go doc'. For per-widget details, run this inside a fugo
project (or the fugo repo) so the fg package resolves; the bare list works
anywhere.

Examples:
  fugo widgets              list the catalog
  fugo widgets Text         show the Text widget's doc and methods
  fugo widgets Button       (also accepts the type name, e.g. ButtonWidget)
  fugo widgets --all        dump the full fg package documentation`,
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:        "all",
				Aliases:     []string{"a"},
				Destination: &all,
				Usage:       "print the full fg package documentation",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, c *cli.Command) error {
			setupUI()

			if all {
				doc, err := runGoDoc(ctx, "-all", fgPkg)
				if err != nil {
					return goDocHint(err)
				}
				out.printf("%s", doc)

				return nil
			}

			if sym := c.Args().First(); sym != "" {
				return showWidget(ctx, sym)
			}

			printWidgetCatalog()

			return nil
		},
	}
}

func printWidgetCatalog() {
	out.heading("Fugo widgets")
	out.printf("  %d widgets in the fg package — constructors are prefix-free: %s, %s.\n\n",
		widgetCount(), out.paint(cCyan, "fg.Text(...)"), out.paint(cCyan, "fg.Button(...)"))

	for _, g := range widgetCatalog {
		out.printf("%s\n", out.paint(cBold, g.name))
		for _, w := range g.items {
			out.printf("  %-22s %s %s\n", out.paint(cCyan, "fg."+w.ctor), out.paint(cDim, "—"), w.desc)
		}
		out.printf("\n")
	}

	out.infof("details: %s    full docs: %s",
		out.paint(cBold, "fugo widgets <name>"), out.paint(cBold, "fugo widgets --all"))
}

func widgetCount() int {
	n := 0
	for _, g := range widgetCatalog {
		n += len(g.items)
	}

	return n
}

// showWidget prints a widget's documentation, preferring the <Name>Widget type
// (which carries the chainable setters) and falling back to the bare symbol.
func showWidget(ctx context.Context, sym string) error {
	for _, target := range []string{sym + "Widget", sym} {
		if doc, err := runGoDoc(ctx, fgPkg, target); err == nil {
			out.printf("%s", doc)

			return nil
		}
	}

	return fmt.Errorf("could not document %q — run inside a fugo project (or the fugo repo) so the fg package resolves; 'fugo widgets' lists the available names", sym)
}

func runGoDoc(ctx context.Context, args ...string) (string, error) {
	cmdArgs := append([]string{"doc"}, args...)
	out.tracef("exec: go %s", strings.Join(cmdArgs, " "))

	o, err := exec.CommandContext(ctx, "go", cmdArgs...).CombinedOutput()

	return string(o), err
}

func goDocHint(err error) error {
	return fmt.Errorf("go doc failed — run inside a fugo project or the fugo repo so the fg package resolves: %w", err)
}
