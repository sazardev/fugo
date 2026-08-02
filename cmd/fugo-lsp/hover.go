package main

import (
	"context"
	"encoding/json"
	"go/types"
	"strings"

	"github.com/sazardev/fugo/lsp"
)

// handleHover implements "textDocument/hover": resolve the identifier under
// the cursor to a types.Object and render its signature (+ doc comment, if
// found) as markdown. Returns nil, nil (not an error) when nothing useful
// resolves at that position — an empty hover is normal LSP behavior, e.g.
// over whitespace or a keyword.
func (s *lspState) handleHover(ctx context.Context, params json.RawMessage) (any, error) {
	var p lsp.HoverParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	path := lsp.FilePath(p.TextDocument.URI)
	content, ok := s.docs.get(p.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	pkg, fset, err := loadPackage(ctx, path, s.docs.overlay())
	if err != nil || pkg == nil {
		return nil, nil
	}

	file := astFileFor(fset, pkg, path)
	if file == nil {
		return nil, nil
	}

	offset := lspPositionToByteOffset(content, p.Position)
	ident, obj := resolveIdentAt(pkg, fset, file, offset)
	if ident == nil || obj == nil {
		return nil, nil
	}

	sig := types.ObjectString(obj, types.RelativeTo(pkg.Types))
	doc := docFor(pkg, obj)

	value := "```go\n" + sig + "\n```"
	if doc != "" {
		value += "\n\n" + strings.TrimSpace(doc)
	}

	identRange := tokenPosToLSPRange(fset, content, ident.Pos(), ident.End())

	return &lsp.Hover{
		Contents: lsp.MarkupContent{Kind: "markdown", Value: value},
		Range:    &identRange,
	}, nil
}
