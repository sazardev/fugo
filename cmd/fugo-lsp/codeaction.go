package main

import (
	"context"
	"encoding/json"

	"github.com/sazardev/fugo/lsp"
)

// handleCodeAction implements "textDocument/codeAction": offers quickfixes
// for any previously-published fugovet diagnostic (see diagnostics.go) that
// overlaps the requested range and carries a SuggestedFix.
func (s *lspState) handleCodeAction(_ context.Context, params json.RawMessage) (any, error) {
	var p lsp.CodeActionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	diags, fset := s.diagsStore.get(p.TextDocument.URI)
	if len(diags) == 0 {
		return []lsp.CodeAction{}, nil
	}

	path := lsp.FilePath(p.TextDocument.URI)
	content := s.contentFor(path)

	var actions []lsp.CodeAction
	for _, d := range diags {
		dRange := tokenPosToLSPRange(fset, content, d.Pos, d.End)
		if !rangesOverlap(dRange, p.Range) {
			continue
		}

		for _, fix := range d.SuggestedFixes {
			edits := make([]lsp.TextEdit, 0, len(fix.TextEdits))
			for _, te := range fix.TextEdits {
				editFile := fset.Position(te.Pos).Filename
				editContent := content
				if lsp.FileURI(editFile) != p.TextDocument.URI {
					editContent = s.contentFor(editFile)
				}
				edits = append(edits, lsp.TextEdit{
					Range:   tokenPosToLSPRange(fset, editContent, te.Pos, te.End),
					NewText: string(te.NewText),
				})
			}

			actions = append(actions, lsp.CodeAction{
				Title: fix.Message,
				Kind:  "quickfix",
				Edit: &lsp.WorkspaceEdit{
					Changes: map[string][]lsp.TextEdit{p.TextDocument.URI: edits},
				},
			})
		}
	}

	return actions, nil
}

// rangesOverlap reports whether a and b share any position.
func rangesOverlap(a, b lsp.Range) bool {
	if lessPos(a.End, b.Start) {
		return false
	}
	if lessPos(b.End, a.Start) {
		return false
	}

	return true
}

func lessPos(a, b lsp.Position) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}

	return a.Character < b.Character
}
