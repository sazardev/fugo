// Package config loads a project's fugo.toml — the declarative source of truth
// for the window (title, size) and the gRPC server address. It is intentionally
// dependency-free (standard library only) so both the fugo runtime and the
// `fugo` CLI can read the same file without pulling in the gRPC stack.
//
// Only the small, fixed schema Fugo emits is supported (a flat key/value subset
// of TOML: comments with '#', '[section]' headers, quoted strings and bare
// integers). It is not a general TOML parser.
package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DefaultName is the conventional config file name in a project root.
const DefaultName = "fugo.toml"

// DefaultAddr is the gRPC listen address used when none is configured.
const DefaultAddr = "127.0.0.1:9510"

// Config is a project's resolved configuration.
type Config struct {
	// Name is the project/binary name (root-level `name = "..."`).
	Name string
	// Window holds the [window] section.
	Window Window
	// Server holds the [server] section.
	Server Server
}

// Window is the [window] section.
type Window struct {
	Title  string
	Width  int
	Height int
}

// Server is the [server] section.
type Server struct {
	Addr string
}

// Default returns the configuration used when no fugo.toml is present.
func Default() Config {
	return Config{
		Name:   "app",
		Window: Window{Title: "Fugo App", Width: 800, Height: 600},
		Server: Server{Addr: DefaultAddr},
	}
}

// Load reads the fugo.toml at path and returns the resolved config, starting
// from Default and overriding with any keys present. If path does not exist it
// returns Default and an error satisfying errors.Is(err, os.ErrNotExist), so
// callers can treat a missing file as "use defaults".
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path) //nolint:gosec // path is a project config file chosen by the caller, not external input
	if err != nil {
		return cfg, err
	}

	if err := parse(data, &cfg); err != nil {
		return cfg, fmt.Errorf("%s: %w", path, err)
	}

	return cfg, nil
}

// Find resolves a config file by name, looking first in the working directory
// and then next to the running executable (so a shipped binary keeps its config
// regardless of the launch directory). It returns the resolved config; a
// missing file yields Default with no error.
func Find(name string) Config {
	if cfg, err := Load(name); err == nil {
		return cfg
	}

	if exe, err := os.Executable(); err == nil {
		beside := filepath.Join(filepath.Dir(exe), name)
		if cfg, err := Load(beside); err == nil {
			return cfg
		}
	}

	return Default()
}

func parse(data []byte, cfg *Config) error {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	section := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])

			continue
		}

		key, raw, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		assign(cfg, section, strings.TrimSpace(key), value(raw))
	}

	return scanner.Err()
}

// value extracts a scalar from the right-hand side of a key/value line: a
// double-quoted string (returned without quotes) or a bare token with any
// trailing inline comment stripped.
func value(raw string) string {
	raw = strings.TrimSpace(raw)

	if strings.HasPrefix(raw, `"`) {
		if end := strings.Index(raw[1:], `"`); end >= 0 {
			return raw[1 : 1+end]
		}

		return strings.Trim(raw, `"`)
	}

	if hash := strings.Index(raw, "#"); hash >= 0 {
		raw = raw[:hash]
	}

	return strings.TrimSpace(raw)
}

func assign(cfg *Config, section, key, val string) {
	switch section {
	case "":
		if key == "name" {
			cfg.Name = val
		}
	case "window":
		switch key {
		case "title":
			cfg.Window.Title = val
		case "width":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.Window.Width = n
			}
		case "height":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.Window.Height = n
			}
		}
	case "server":
		if key == "addr" {
			cfg.Server.Addr = val
		}
	}
}
