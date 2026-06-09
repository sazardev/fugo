package fugo

import (
	"os"

	"github.com/sazardev/fugo/config"
)

// ConfigOptions reads a project's config file (typically "fugo.toml"), looking
// in the working directory and then beside the running executable, and returns
// the window AppOptions. The configured server address is exported as FUGO_ADDR
// unless one is already set, so an explicit env or `fugo run --addr` still wins.
// A missing or unreadable file falls back to built-in defaults, so a generated
// app runs even if the file is deleted.
//
// Generated projects call it as:
//
//	fugo.RunStandalone(fugo.ConfigOptions("fugo.toml"), ui.Build)
func ConfigOptions(name string) AppOptions {
	cfg := config.Find(name)

	if cfg.Server.Addr != "" && os.Getenv("FUGO_ADDR") == "" {
		_ = os.Setenv("FUGO_ADDR", cfg.Server.Addr)
	}

	return AppOptions{
		Title:  cfg.Window.Title,
		Width:  cfg.Window.Width,
		Height: cfg.Window.Height,
	}
}
