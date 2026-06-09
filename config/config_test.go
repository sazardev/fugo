package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sazardev/fugo/config"
)

func TestLoadFull(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "fugo.toml")
	body := `# project config
name = "demo"

[window]
title  = "My Demo"   # inline comment
width  = 1024
height = 720

[server]
addr = "127.0.0.1:9600"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Name != "demo" {
		t.Errorf("name = %q, want demo", cfg.Name)
	}
	if cfg.Window.Title != "My Demo" {
		t.Errorf("title = %q, want My Demo", cfg.Window.Title)
	}
	if cfg.Window.Width != 1024 || cfg.Window.Height != 720 {
		t.Errorf("size = %dx%d, want 1024x720", cfg.Window.Width, cfg.Window.Height)
	}
	if cfg.Server.Addr != "127.0.0.1:9600" {
		t.Errorf("addr = %q, want 127.0.0.1:9600", cfg.Server.Addr)
	}
}

func TestLoadPartialKeepsDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "fugo.toml")
	if err := os.WriteFile(path, []byte("[window]\ntitle = \"Only Title\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	def := config.Default()
	if cfg.Window.Title != "Only Title" {
		t.Errorf("title = %q, want Only Title", cfg.Window.Title)
	}
	if cfg.Window.Width != def.Window.Width || cfg.Window.Height != def.Window.Height {
		t.Errorf("size = %dx%d, want defaults %dx%d", cfg.Window.Width, cfg.Window.Height, def.Window.Width, def.Window.Height)
	}
	if cfg.Server.Addr != def.Server.Addr {
		t.Errorf("addr = %q, want default %q", cfg.Server.Addr, def.Server.Addr)
	}
}

func TestLoadMissingIsNotExist(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(filepath.Join(t.TempDir(), "absent.toml"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v, want os.ErrNotExist", err)
	}
	if cfg != config.Default() {
		t.Errorf("cfg = %+v, want Default on missing file", cfg)
	}
}

func TestLoadIgnoresBadInts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "fugo.toml")
	if err := os.WriteFile(path, []byte("[window]\nwidth = not-a-number\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Window.Width != config.Default().Window.Width {
		t.Errorf("width = %d, want default kept on bad int", cfg.Window.Width)
	}
}

func TestFindFallsBackToDefault(t *testing.T) {
	t.Parallel()

	// A name unlikely to exist in CWD or beside the test binary.
	cfg := config.Find("definitely-not-here-fugo.toml")
	if cfg != config.Default() {
		t.Errorf("Find = %+v, want Default", cfg)
	}
}
