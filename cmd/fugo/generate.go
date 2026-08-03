package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli/v3"
)

// generateCmd scaffolds boilerplate inside an existing project's ui package,
// so adding a screen or a reusable component doesn't start from a blank file.
func generateCmd() *cli.Command {
	return &cli.Command{
		Name:  "generate",
		Usage: "Scaffold a new screen or reusable component in the ui package",
		Description: `Generate boilerplate Go files inside the ui package of an existing Fugo
project. Run this from the project root (where main.go lives).

Examples:
  fugo generate screen Settings           ui/settings.go with a SettingsPage(ctx) screen
  fugo generate component PrimaryButton   ui/primary_button.go with a reusable fg.Widget helper`,
		Commands: []*cli.Command{
			generateScreenCmd(),
			generateComponentCmd(),
		},
	}
}

func generateScreenCmd() *cli.Command {
	var force bool

	return &cli.Command{
		Name:      "screen",
		Usage:     "Scaffold a new screen (a <Name>Page function) in the ui package",
		ArgsUsage: "<Name>",
		Description: `Writes ui/<name>.go with a <Name>Page(ctx *fugo.Context) fg.Widget function
and a comment showing how to wire it into the Router in ui.Build.`,
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Destination: &force,
				Usage:       "overwrite the file if it already exists",
			},
		}, verbosityFlags()...),
		Action: func(_ context.Context, c *cli.Command) error {
			setupUI()

			name := c.Args().First()
			if name == "" {
				return errors.New("screen name required: fugo generate screen <Name>")
			}

			return generateScreen(name, force)
		},
	}
}

func generateComponentCmd() *cli.Command {
	var force bool

	return &cli.Command{
		Name:      "component",
		Usage:     "Scaffold a reusable component (a fg.Widget-returning function) in the ui package",
		ArgsUsage: "<Name>",
		Description: `Writes ui/<name>.go with a %s(t fg.Theme) fg.Widget function — a plain Go
helper you compose into your screens, not a new fg widget type.`,
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Destination: &force,
				Usage:       "overwrite the file if it already exists",
			},
		}, verbosityFlags()...),
		Action: func(_ context.Context, c *cli.Command) error {
			setupUI()

			name := c.Args().First()
			if name == "" {
				return errors.New("component name required: fugo generate component <Name>")
			}

			return generateComponent(name, force)
		},
	}
}

func generateScreen(rawName string, force bool) error {
	dir, err := uiPackageDir()
	if err != nil {
		return err
	}

	base := identForName(rawName)
	if base == "" {
		return fmt.Errorf("invalid screen name %q", rawName)
	}

	funcName := base + "Page"
	file := filepath.Join(dir, fileForName(base)+".go")
	if err := checkNotExists(file, force); err != nil {
		return err
	}

	route := "/" + strings.ReplaceAll(fileForName(base), "_", "-")
	content := fmt.Sprintf(screenFileTemplate, funcName, route, base)

	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", file, err)
	}

	out.successf("generated %s", file)
	out.infof("wire it in: add %q: func() fg.Widget { return %s(ctx) } to the map passed to fg.Router in ui.Build",
		route, funcName)

	return nil
}

func generateComponent(rawName string, force bool) error {
	dir, err := uiPackageDir()
	if err != nil {
		return err
	}

	name := identForName(rawName)
	if name == "" {
		return fmt.Errorf("invalid component name %q", rawName)
	}

	file := filepath.Join(dir, fileForName(name)+".go")
	if err := checkNotExists(file, force); err != nil {
		return err
	}

	content := fmt.Sprintf(componentFileTemplate, name)

	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", file, err)
	}

	out.successf("generated %s", file)

	return nil
}

func checkNotExists(file string, force bool) error {
	if _, err := os.Stat(file); err == nil && !force {
		return fmt.Errorf("%s already exists — pass --force to overwrite", file)
	}

	return nil
}

// uiPackageDir locates the project's ui package: the directory imported by
// main.go's `<module>/ui`-suffixed import, or a bare ./ui if main.go doesn't
// import one but the directory exists.
func uiPackageDir() (string, error) {
	if !hasMainGo() {
		return "", errors.New("no main.go in the current directory — run this inside a fugo project")
	}

	module := goModModule("go.mod")
	if module == "" {
		return "", errors.New("no go.mod in the current directory — run this inside a fugo project")
	}

	for _, imp := range goImports("main.go") {
		if !strings.HasSuffix(imp, "/ui") {
			continue
		}
		if !strings.HasPrefix(imp, module+"/") {
			return "", fmt.Errorf("main.go imports %q but go.mod module is %q — run 'fugo doctor' to reconcile", imp, module)
		}

		return strings.TrimPrefix(imp, module+"/"), nil
	}

	if info, err := os.Stat("ui"); err == nil && info.IsDir() {
		return "ui", nil
	}

	return "", errors.New("could not find a ui package — main.go doesn't import one and ./ui doesn't exist")
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]+`)

// identForName turns arbitrary user input ("user profile", "user-profile")
// into an exported Go identifier ("UserProfile").
func identForName(name string) string {
	var b strings.Builder
	for _, part := range nonAlnum.Split(name, -1) {
		if part == "" {
			continue
		}

		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}

	return b.String()
}

// fileForName turns an identifier ("UserProfile") into a snake_case file stem
// ("user_profile").
func fileForName(ident string) string {
	var b strings.Builder
	for i, r := range ident {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}

		b.WriteRune(r)
	}

	return strings.ToLower(b.String())
}

const screenFileTemplate = `package ui

import (
	"github.com/sazardev/fugo"
	"github.com/sazardev/fugo/fg"
)

// %[1]s is a screen — wire it into the Router in Build (ui.Build). Example:
//
//	%[2]q: func() fg.Widget { return %[1]s(ctx) },
func %[1]s(ctx *fugo.Context) fg.Widget {
	_ = ctx // remove once you wire up event handlers

	t := fg.CurrentTheme()

	return fg.Container(fg.Column(
		fg.Text("%[3]s").FontSize(t.Typography.Heading).Weight(fg.WeightBold),
		fg.Divider().Color(t.Colors.Border),
		fg.SizedBox(0, t.Spacing.MD),
		fg.Text("New screen — build it out here.").Color(t.Colors.Muted),
	)).
		BgColor(t.Colors.Background).
		Pad(fg.EdgeAll(t.Spacing.LG))
}
`

const componentFileTemplate = `package ui

import (
	"github.com/sazardev/fugo/fg"
)

// %[1]s is a reusable component — compose it into your screens.
func %[1]s(t fg.Theme) fg.Widget {
	return fg.Container(
		fg.Text("%[1]s"),
	).
		BgColor(t.Colors.Surface).
		Pad(fg.EdgeAll(t.Spacing.MD))
}
`
