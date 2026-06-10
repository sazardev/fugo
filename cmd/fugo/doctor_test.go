package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoModModule(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(path, []byte("module github.com/me/app\n\ngo 1.26.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if got := goModModule(path); got != "github.com/me/app" {
		t.Errorf("goModModule = %q, want github.com/me/app", got)
	}
	if got := goModModule(filepath.Join(dir, "absent.mod")); got != "" {
		t.Errorf("goModModule(absent) = %q, want empty", got)
	}
}

func TestGoImports(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	src := `package main

import (
	"strconv"

	"github.com/sazardev/fugo"

	"myapp/ui"
)

var _ = strconv.Itoa
`
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}

	imports := goImports(path)
	want := map[string]bool{"strconv": true, "github.com/sazardev/fugo": true, "myapp/ui": true}
	if len(imports) != len(want) {
		t.Fatalf("imports = %v, want %d entries", imports, len(want))
	}
	for _, imp := range imports {
		if !want[imp] {
			t.Errorf("unexpected import %q", imp)
		}
	}
}

func TestHasBuildFunc(t *testing.T) {
	t.Parallel()

	withBuild := t.TempDir()
	if err := os.WriteFile(filepath.Join(withBuild, "home.go"),
		[]byte("package ui\n\ntype C struct{}\n\nfunc Build(c *C) int { return 0 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !hasBuildFunc(withBuild) {
		t.Error("hasBuildFunc = false, want true (Build is declared)")
	}

	noBuild := t.TempDir()
	if err := os.WriteFile(filepath.Join(noBuild, "home.go"),
		[]byte("package ui\n\nfunc Render() int { return 0 }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if hasBuildFunc(noBuild) {
		t.Error("hasBuildFunc = true, want false (no Build func)")
	}

	// A method named Build (with a receiver) must NOT count.
	method := t.TempDir()
	if err := os.WriteFile(filepath.Join(method, "home.go"),
		[]byte("package ui\n\ntype T struct{}\n\nfunc (T) Build() {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if hasBuildFunc(method) {
		t.Error("hasBuildFunc = true, want false (Build is a method, not a package func)")
	}
}

func TestInProject(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	if inProject() {
		t.Error("inProject = true in an empty dir, want false")
	}

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !inProject() {
		t.Error("inProject = false with main.go present, want true")
	}
}
