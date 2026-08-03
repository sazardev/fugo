package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/sazardev/fugo/lsp"
)

func TestLSPPositionToByteOffset_ASCII(t *testing.T) {
	content := "line one\nline two\nline three"

	tests := []struct {
		name string
		pos  lsp.Position
		want int
	}{
		{"start", lsp.Position{Line: 0, Character: 0}, 0},
		{"mid line 0", lsp.Position{Line: 0, Character: 4}, 4},
		{"start line 1", lsp.Position{Line: 1, Character: 0}, 9},
		{"mid line 1", lsp.Position{Line: 1, Character: 5}, 14},
		{"start line 2", lsp.Position{Line: 2, Character: 0}, 18},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lspPositionToByteOffset(content, tt.pos)
			if got != tt.want {
				t.Errorf("lspPositionToByteOffset(%q, %+v) = %d, want %d", content, tt.pos, got, tt.want)
			}
		})
	}
}

func TestLSPPositionToByteOffset_MultiByte(t *testing.T) {
	// "→" is U+2192, 3 bytes in UTF-8, 1 code unit in UTF-16 (within the
	// BMP). Confirm characters AFTER it on the line are offset correctly.
	line := "About →Home"
	content := line

	// "About " is 6 ASCII chars/bytes/UTF-16 units. "→" occupies UTF-16
	// character index 6 (1 unit), byte offset 6 (3 bytes long).
	// "Home" starts at UTF-16 character 7, byte offset 6+3=9.
	arrowByteOffset := lspPositionToByteOffset(content, lsp.Position{Line: 0, Character: 6})
	if arrowByteOffset != 6 {
		t.Errorf("byte offset of arrow = %d, want 6", arrowByteOffset)
	}

	afterArrow := lspPositionToByteOffset(content, lsp.Position{Line: 0, Character: 7})
	if afterArrow != 9 {
		t.Errorf("byte offset after arrow = %d, want 9 (got substring %q)", afterArrow, content[afterArrow:])
	}
	if content[afterArrow:] != "Home" {
		t.Errorf("content after arrow = %q, want %q", content[afterArrow:], "Home")
	}
}

func TestByteOffsetToLSPPosition_RoundTrip(t *testing.T) {
	contents := []string{
		"line one\nline two\nline three",
		"About →Home\n← Back\nplain ascii",
		"single line, no newline",
		"",
		"\n\n\n",
	}

	for _, content := range contents {
		// Only check offsets that land on a rune boundary — an offset
		// mid-way through a multi-byte UTF-8 sequence has no meaningful
		// LSP position of its own, so round-tripping it isn't expected to
		// return the same byte offset.
		for offset := range content {
			pos := byteOffsetToLSPPosition(content, offset)
			back := lspPositionToByteOffset(content, pos)
			if back != offset {
				t.Errorf("round-trip failed for content %q at offset %d: got position %+v -> offset %d", content, offset, pos, back)
			}
		}
		// Also check the end-of-content offset explicitly.
		pos := byteOffsetToLSPPosition(content, len(content))
		back := lspPositionToByteOffset(content, pos)
		if back != len(content) {
			t.Errorf("round-trip failed for content %q at end offset %d: got position %+v -> offset %d", content, len(content), pos, back)
		}
	}
}

func TestTokenPosToLSPRange(t *testing.T) {
	src := `package p

// About → does the thing.
func About() {
	x := 1
	_ = x
}
`
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	var fn *ast.FuncDecl
	ast.Inspect(astFile, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == "About" {
			fn = fd

			return false
		}

		return true
	})
	if fn == nil {
		t.Fatal("could not find func About in parsed source")
	}

	r := tokenPosToLSPRange(fset, src, fn.Name.Pos(), fn.Name.End())
	if lessPos(r.End, r.Start) {
		t.Fatalf("range end before start: %+v", r)
	}

	// Verify the range actually covers "About" by re-deriving byte offsets
	// from it and slicing src.
	startOff := lspPositionToByteOffset(src, r.Start)
	endOff := lspPositionToByteOffset(src, r.End)
	if got := src[startOff:endOff]; got != "About" {
		t.Errorf("range covers %q, want %q", got, "About")
	}
}
