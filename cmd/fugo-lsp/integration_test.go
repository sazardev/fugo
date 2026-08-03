package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sazardev/fugo/lsp"
)

// writePlainFixture creates a minimal, dependency-free Go module in a temp
// dir (no import of fg) for testing hover/definition against plain
// identifiers, without needing network access or the fugo module's own
// dependency graph. Returns the fixture's main.go path.
func writePlainFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	goMod := "module fixture\n\ngo 1.26\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	src := `package main

// Greeting is the message shown to the user.
const Greeting = "hello"

// Greet returns a friendly greeting for name.
func Greet(name string) string {
	return Greeting + ", " + name
}

func main() {
	msg := Greet("world")
	_ = msg
}
`
	filePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filePath, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	return filePath
}

func newTestState() *lspState {
	return newLSPState(lsp.NewServer())
}

func TestHover_ResolvesFunction(t *testing.T) {
	filePath := writePlainFixture(t)
	src, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	s := newTestState()
	uri := lsp.FileURI(filePath)
	s.docs.set(uri, string(src))

	// "Greet" appears in "msg := Greet(\"world\")" — find it.
	content := string(src)
	idx := indexOf(content, `Greet("world")`)
	if idx < 0 {
		t.Fatal("fixture missing expected call site")
	}
	pos := byteOffsetToLSPPosition(content, idx+1) // land inside "Greet"

	pkg, fset, err := loadPackage(context.Background(), filePath, s.docs.overlay())
	if err != nil {
		t.Fatalf("loadPackage: %v", err)
	}
	file := astFileFor(fset, pkg, filePath)
	if file == nil {
		t.Fatal("astFileFor returned nil")
	}

	offset := lspPositionToByteOffset(content, pos)
	ident, obj := resolveIdentAt(pkg, fset, file, offset)
	if ident == nil || obj == nil {
		t.Fatalf("resolveIdentAt found nothing at offset %d (pos %+v)", offset, pos)
	}
	if ident.Name != "Greet" {
		t.Fatalf("resolved ident = %q, want Greet", ident.Name)
	}

	doc := docFor(pkg, obj)
	if doc == "" {
		t.Error("expected a doc comment for Greet, got none")
	}
}

func TestDefinition_PointsAtDeclaration(t *testing.T) {
	filePath := writePlainFixture(t)
	src, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(src)

	s := newTestState()
	uri := lsp.FileURI(filePath)
	s.docs.set(uri, content)

	idx := indexOf(content, `Greet("world")`)
	pos := byteOffsetToLSPPosition(content, idx+1)

	params := mustMarshal(t, lsp.DefinitionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Position:     pos,
		},
	})

	result, err := s.handleDefinition(context.Background(), params)
	if err != nil {
		t.Fatalf("handleDefinition: %v", err)
	}
	locs, ok := result.([]lsp.Location)
	if !ok || len(locs) == 0 {
		t.Fatalf("handleDefinition returned %#v, want a non-empty []lsp.Location", result)
	}
	if locs[0].URI != uri {
		t.Errorf("definition URI = %q, want %q", locs[0].URI, uri)
	}

	declOffset := lspPositionToByteOffset(content, locs[0].Range.Start)
	if got := content[declOffset : declOffset+5]; got != "Greet" {
		t.Errorf("definition points at %q, want \"Greet\"", got)
	}
}

func TestCompletion_FgPackageCatalog(t *testing.T) {
	if testing.Short() {
		t.Skip("requires module resolution against the local fugo module")
	}

	dir := t.TempDir()

	goMod := `module fixture

go 1.26

require github.com/sazardev/fugo v0.0.0

replace github.com/sazardev/fugo => ` + repoRoot(t) + `
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		t.Fatal(err)
	}

	src := `package main

import "github.com/sazardev/fugo/fg"

func main() {
	fg.Button("hi")
}
`
	filePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filePath, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	// go mod tidy so the module graph resolves against the replace target.
	cmd := exec.CommandContext(context.Background(), "go", "mod", "tidy")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy: %v\n%s", err, out)
	}

	s := newTestState()
	uri := lsp.FileURI(filePath)
	s.docs.set(uri, src)

	// Position right after "fg." on the line "\tfg.Button(\"hi\")\n".
	dotIdx := indexOf(src, "fg.Button")
	pos := byteOffsetToLSPPosition(src, dotIdx+3) // just after "fg."

	params := mustMarshal(t, lsp.CompletionParams{
		TextDocumentPositionParams: lsp.TextDocumentPositionParams{
			TextDocument: lsp.TextDocumentIdentifier{URI: uri},
			Position:     pos,
		},
	})

	result, err := s.handleCompletion(context.Background(), params)
	if err != nil {
		t.Fatalf("handleCompletion: %v", err)
	}
	list, ok := result.(lsp.CompletionList)
	if !ok {
		t.Fatalf("handleCompletion returned %#v, not lsp.CompletionList", result)
	}
	if len(list.Items) == 0 {
		t.Fatal("completion returned no items")
	}

	want := map[string]bool{"Button": false, "Text": false, "Container": false}
	for _, item := range list.Items {
		if _, ok := want[item.Label]; ok {
			want[item.Label] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("completion items missing %q (got %d items)", name, len(list.Items))
		}
	}
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// cmd/fugo-lsp -> repo root is two levels up.
	return filepath.Dir(filepath.Dir(wd))
}
