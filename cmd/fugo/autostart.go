package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli/v3"
)

// autostartCmd toggles launching the current project's built app at login —
// the realistic shape of "run in the background" for a desktop GUI app (a
// systemd/Windows *service* has no window to show; autostart is what users
// actually mean).
func autostartCmd() *cli.Command {
	return &cli.Command{
		Name:  "autostart",
		Usage: "Enable or disable launching this app at login",
		Description: `Registers (or removes) the built app in dist/ to launch automatically when
you log in — on Linux via an XDG autostart .desktop entry
(~/.config/autostart/), on Windows via a Run registry key. Run 'fugo build'
first; autostart points at the binary in dist/.

Linux desktop-environment note: GNOME, KDE, XFCE, and similar session
managers process ~/.config/autostart/ automatically. Tiling Wayland
compositors without a full session manager (Hyprland, Sway, ...) generally do
not — you also need something like 'dex' invoked from your compositor config
(e.g. Hyprland's "exec-once = dex ~/.config/autostart -a") for the entry to
actually run.

Examples:
  fugo autostart enable
  fugo autostart disable`,
		Commands: []*cli.Command{
			{
				Name:  "enable",
				Usage: "Launch this app at login",
				Flags: verbosityFlags(),
				Action: func(ctx context.Context, _ *cli.Command) error {
					setupUI()

					return setAutostart(ctx, true)
				},
			},
			{
				Name:  "disable",
				Usage: "Stop launching this app at login",
				Flags: verbosityFlags(),
				Action: func(ctx context.Context, _ *cli.Command) error {
					setupUI()

					return setAutostart(ctx, false)
				},
			},
		},
	}
}

// setAutostart writes or removes the platform-specific autostart entry for
// the current project's built binary.
func setAutostart(ctx context.Context, enable bool) error {
	name := projectName()

	binPath, err := filepath.Abs(filepath.Join(distDir, name+exeSuffix()))
	if err != nil {
		return err
	}

	if _, statErr := os.Stat(binPath); statErr != nil {
		return fmt.Errorf("%s not found — run 'fugo build' first", binPath)
	}

	if runtime.GOOS == osWindows {
		return setAutostartWindows(ctx, name, binPath, enable)
	}

	return setAutostartLinux(name, binPath, enable)
}

// setAutostartLinux writes (or removes) an XDG autostart .desktop entry.
func setAutostartLinux(name, binPath string, enable bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(home, ".config", "autostart")
	path := filepath.Join(dir, name+".desktop")

	if !enable {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}

		out.successf("removed %s", path)

		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	entry := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s
Terminal=false
X-GNOME-Autostart-enabled=true
`, name, binPath)

	if err := os.WriteFile(path, []byte(entry), 0o644); err != nil {
		return err
	}

	out.successf("wrote %s", path)
	out.infof("  Hyprland/Sway/other session-manager-less compositors need 'dex' (or similar) to actually run entries in ~/.config/autostart/ — see 'fugo autostart --help'")

	return nil
}

// setAutostartWindows adds (or removes) a per-user Run registry key via reg.exe,
// avoiding a direct registry-package dependency for a single CLI feature.
func setAutostartWindows(ctx context.Context, name, binPath string, enable bool) error {
	const runKey = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`

	var cmd *exec.Cmd
	if enable {
		cmd = exec.CommandContext(ctx, "reg", "add", runKey, "/v", name, "/t", "REG_SZ", "/d", binPath, "/f")
	} else {
		cmd = exec.CommandContext(ctx, "reg", "delete", runKey, "/v", name, "/f")
	}

	regOut, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("reg: %w: %s", err, string(regOut))
	}

	if enable {
		out.successf("added %s to %s", name, runKey)
	} else {
		out.successf("removed %s from %s", name, runKey)
	}

	return nil
}
