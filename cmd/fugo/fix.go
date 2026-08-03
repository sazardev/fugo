package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

// vetCmd runs the fugovet analyzer (Fugo-specific static checks) over the
// project, on top of whatever `go vet` and golangci-lint already cover.
func vetCmd() *cli.Command {
	return &cli.Command{
		Name:  "vet",
		Usage: "Run Fugo's opinionated static analyzer over the project",
		Description: `Checks Fugo-specific conventions that go vet doesn't know about: a widget
setter call (SetText, SetValue, ...) with no ctx.Update()/ctx.UpdateNow()
reachable in the same function, a Button-family constructor with no OnClick
handler, and a ui import with no exported Build function.

'fugo fix' applies the mechanical subset of these automatically.`,
		Flags: verbosityFlags(),
		Action: func(ctx context.Context, _ *cli.Command) error {
			setupUI()

			bin, err := ensureFugovet(ctx)
			if err != nil {
				return err
			}

			cmd := exec.CommandContext(ctx, bin, "./...")

			return out.runStep("Running the fugo analyzer", cmd)
		},
	}
}

// fixCmd auto-fixes what it safely can, then applies Fugo's opinionated
// declaration ordering.
func fixCmd() *cli.Command {
	var noReorder bool

	return &cli.Command{
		Name:  "fix",
		Usage: "Auto-fix analyzer findings, format, and reorder declarations by Fugo convention",
		Description: `Three passes over the project:

  1. fugovet -fix   mechanical fixes only (e.g. inserting a missing ctx.Update())
  2. gofumpt -w .   the project's formatter
  3. reorder        (unless --no-reorder) within each file: groups top-level
                    funcs so entrypoints (main/Build) come first, then other
                    exported funcs, then unexported ones — each group keeping
                    its original relative order — and sorts fg.Router route
                    maps alphabetically by path. gofumpt runs once more after
                    to clean up the resulting spacing.

Fugo is opinionated about layout on purpose: a file organized this way reads
the same way in every project. Review the diff before committing — pass
--no-reorder to skip step 3 if you'd rather keep your own ordering.`,
		Flags: append([]cli.Flag{
			&cli.BoolFlag{
				Name:        "no-reorder",
				Destination: &noReorder,
				Usage:       "skip the structural reordering pass",
			},
		}, verbosityFlags()...),
		Action: func(ctx context.Context, _ *cli.Command) error {
			setupUI()

			bin, err := ensureFugovet(ctx)
			if err != nil {
				return err
			}

			vetFix := exec.CommandContext(ctx, bin, "-fix", "./...")
			if err := out.runStep("Applying mechanical fixes (fugovet -fix)", vetFix); err != nil {
				out.warnf("some findings need a human — run 'fugo vet' for details")
			}

			if err := gofumptProject(ctx); err != nil {
				return err
			}

			if !noReorder {
				n, err := reorderProject(".")
				if err != nil {
					return fmt.Errorf("reorder: %w", err)
				}
				if n > 0 {
					out.successf("reordered %d file(s)", n)

					if err := gofumptProject(ctx); err != nil {
						return err
					}
				}
			}

			build := exec.CommandContext(ctx, "go", subcmdBuild, "./...")
			if err := out.runStep("Verifying (go build ./...)", build); err != nil {
				out.warnf("the project no longer builds after fixing — review the changes (git diff) before committing")

				return err
			}

			out.successf("fixed")

			return nil
		},
	}
}

func gofumptProject(ctx context.Context) error {
	if _, err := exec.LookPath("gofumpt"); err != nil {
		out.warnf("gofumpt not on PATH — skipping formatting (install: go install mvdan.cc/gofumpt@latest)")

		return nil //nolint:nilerr // intentional: a missing optional tool degrades to a warning, not a failed 'fugo fix'
	}

	cmd := exec.CommandContext(ctx, "gofumpt", "-w", ".")

	return out.runStep("Formatting (gofumpt)", cmd)
}

// ensureFugovet builds the fugovet analyzer binary fresh from the resolved
// fugo module source (the same module the current project already depends
// on — honoring any replace directive, exactly like fugoModuleDir is used
// elsewhere), so it always matches the fugo version in use.
func ensureFugovet(ctx context.Context) (string, error) {
	repo := fugoModuleDir(ctx)
	if repo == "" {
		return "", errors.New("could not resolve the fugo module — run this inside a fugo project (go.mod must require github.com/sazardev/fugo)")
	}

	bin := filepath.Join(os.TempDir(), "fugovet"+exeSuffix())

	build := exec.CommandContext(ctx, "go", subcmdBuild, "-o", bin, "./cmd/fugovet")
	build.Dir = repo
	if err := out.runStep("Building the fugo analyzer", build); err != nil {
		return "", fmt.Errorf("build fugovet: %w", err)
	}

	return bin, nil
}
