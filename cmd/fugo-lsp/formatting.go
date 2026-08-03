package main

import (
	"bytes"
	"context"
	"encoding/json"
	"go/format"
	"os/exec"

	"github.com/sazardev/fugo/lsp"
)

// handleFormatting implements "textDocument/formatting": formats the
// current buffer content (not disk content) with gofumpt if it's on PATH,
// falling back to go/format.Source otherwise. Returns nil (no edits, not an
// error) if the content can't be formatted (e.g. a syntax error) — a failed
// format should just be a no-op, not a request failure.
func (s *lspState) handleFormatting(ctx context.Context, params json.RawMessage) (any, error) {
	var p lsp.DocumentFormattingParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	content, ok := s.docs.get(p.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	formatted, ok := formatSource(ctx, content)
	if !ok {
		return nil, nil
	}
	if formatted == content {
		return []lsp.TextEdit{}, nil
	}

	end := byteOffsetToLSPPosition(content, len(content))

	return []lsp.TextEdit{{
		Range:   lsp.Range{Start: lsp.Position{Line: 0, Character: 0}, End: end},
		NewText: formatted,
	}}, nil
}

// formatSource formats src, preferring the gofumpt binary if it's on PATH
// (Fugo's canonical formatter, per CLAUDE.md), falling back to
// go/format.Source. Returns ok=false if formatting failed (syntax error).
func formatSource(ctx context.Context, src string) (string, bool) {
	if binPath, err := exec.LookPath("gofumpt"); err == nil {
		cmd := exec.CommandContext(ctx, binPath)
		cmd.Stdin = bytes.NewReader([]byte(src))
		out, err := cmd.Output()
		if err == nil {
			return string(out), true
		}
		// Fall through to go/format on gofumpt failure (e.g. syntax error
		// during an in-progress edit) — still worth trying the simpler
		// formatter in case it's more forgiving, otherwise both will fail
		// identically and we report no edit.
	}

	out, err := format.Source([]byte(src))
	if err != nil {
		return "", false
	}

	return string(out), true
}
