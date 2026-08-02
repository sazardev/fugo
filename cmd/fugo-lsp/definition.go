package main

import (
	"context"
	"encoding/json"
	"go/token"
	"os"

	"github.com/sazardev/fugo/lsp"
)

// handleDefinition implements "textDocument/definition": resolve the
// identifier under the cursor and point at its declaration. Returns nil,
// nil (not an error) when nothing resolves.
func (s *lspState) handleDefinition(ctx context.Context, params json.RawMessage) (any, error) {
	var p lsp.DefinitionParams
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
	_, obj := resolveIdentAt(pkg, fset, file, offset)
	if obj == nil || !obj.Pos().IsValid() {
		return nil, nil
	}

	declPos := fset.Position(obj.Pos())
	declFilename := declPos.Filename
	declEnd := obj.Pos() + token.Pos(declEndOffset(obj))

	// The declaring file may be different from the one hover/definition was
	// requested in (e.g. jumping into a dependency, or another file in the
	// same package) — its content may or may not be in the overlay.
	declContent, ok := s.docs.get(lsp.FileURI(declFilename))
	if !ok {
		raw, err := os.ReadFile(declFilename)
		if err != nil {
			return nil, nil
		}
		declContent = string(raw)
	}

	r := tokenPosToLSPRange(fset, declContent, obj.Pos(), declEnd)

	return []lsp.Location{{
		URI:   lsp.FileURI(declFilename),
		Range: r,
	}}, nil
}

// declEndOffset returns how many bytes past obj.Pos() its identifier
// spans, so the range covers at least the declared name (types.Object
// doesn't expose an End(), only Pos(), so this approximates it from the
// object's name length).
func declEndOffset(obj interface{ Name() string }) int {
	return len(obj.Name())
}
