package main

import (
	"go/token"
	"strings"
	"unicode/utf8"

	"github.com/sazardev/fugo/lsp"
)

// splitLines splits content into lines using "\n" as the separator, keeping
// the trailing empty "line" if content ends with "\n" (so line indices line
// up the same way they would in a text editor buffer).
func splitLines(content string) []string {
	return strings.Split(content, "\n")
}

// lspPositionToByteOffset converts an lsp.Position (0-indexed line, UTF-16
// code-unit character offset within that line) into an absolute byte offset
// into content. content is assumed to use "\n" line separators.
//
// If pos.Line is beyond the end of content, the offset returned is
// len(content). If pos.Character is beyond the end of its line (in UTF-16
// units), the offset returned is the byte offset of the end of that line.
func lspPositionToByteOffset(content string, pos lsp.Position) int {
	lines := splitLines(content)

	line := int(pos.Line)
	if line >= len(lines) {
		return len(content)
	}

	// Byte offset of the start of `line`: sum of all previous lines' byte
	// lengths plus one '\n' each.
	offset := 0
	for i := 0; i < line; i++ {
		offset += len(lines[i]) + 1 // +1 for the '\n' separator
	}

	offset += utf16OffsetToByteOffset(lines[line], int(pos.Character))

	return offset
}

// utf16OffsetToByteOffset converts a UTF-16 code-unit offset within a single
// line (no newlines) into a byte offset within that line. If the requested
// offset is beyond the end of the line (in UTF-16 units), the line's full
// byte length is returned.
func utf16OffsetToByteOffset(line string, utf16Offset int) int {
	if utf16Offset <= 0 {
		return 0
	}

	units := 0
	byteOff := 0
	for _, r := range line {
		if units >= utf16Offset {
			break
		}
		byteOff += utf8.RuneLen(r)
		units += utf16RuneLen(r)
	}
	if units < utf16Offset {
		// Requested offset lands beyond the line's content (e.g. exactly
		// at or past the end); clamp to the line's byte length.
		return len(line)
	}

	return byteOff
}

// byteOffsetToLSPPosition is the inverse of lspPositionToByteOffset: given
// an absolute byte offset into content, computes the corresponding
// lsp.Position (0-indexed line, UTF-16 code-unit character offset).
func byteOffsetToLSPPosition(content string, offset int) lsp.Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(content) {
		offset = len(content)
	}

	// Find the line containing offset by counting '\n' bytes before it.
	line := 0
	lineStart := 0
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	lineContent := content[lineStart:offset]
	character := 0
	for _, r := range lineContent {
		character += utf16RuneLen(r)
	}

	return lsp.Position{Line: uint32(line), Character: uint32(character)}
}

// utf16RuneLen returns how many UTF-16 code units r encodes to: 1 for any
// rune within the Basic Multilingual Plane (<= 0xFFFF, which includes
// characters like '→' U+2192), 2 for supplementary-plane runes (e.g. most
// emoji), via a surrogate pair.
func utf16RuneLen(r rune) int {
	if r > 0xFFFF {
		return 2
	}

	return 1
}

// tokenPosToLSPRange converts a [start, end) token.Pos range (as produced by
// go/ast, go/types, go/analysis, ...) into an lsp.Range, using fset to
// resolve positions to file/line/column and content as the authoritative
// source text for that file (which may differ from what's on disk if the
// buffer has unsaved edits).
func tokenPosToLSPRange(fset *token.FileSet, content string, start, end token.Pos) lsp.Range {
	startPos := fset.Position(start)
	endPos := fset.Position(end)

	return lsp.Range{
		Start: lineColToLSPPosition(content, startPos.Line, startPos.Column),
		End:   lineColToLSPPosition(content, endPos.Line, endPos.Column),
	}
}

// lineColToLSPPosition converts a go/token 1-indexed line + 1-indexed
// byte-column pair into an lsp.Position (0-indexed line + UTF-16 character).
func lineColToLSPPosition(content string, line, col int) lsp.Position {
	lines := splitLines(content)
	lineIdx := line - 1
	if lineIdx < 0 {
		lineIdx = 0
	}
	if lineIdx >= len(lines) {
		return byteOffsetToLSPPosition(content, len(content))
	}

	lineContent := lines[lineIdx]
	byteCol := col - 1
	if byteCol < 0 {
		byteCol = 0
	}
	if byteCol > len(lineContent) {
		byteCol = len(lineContent)
	}

	character := 0
	for _, r := range lineContent[:byteCol] {
		character += utf16RuneLen(r)
	}

	return lsp.Position{Line: uint32(lineIdx), Character: uint32(character)}
}
