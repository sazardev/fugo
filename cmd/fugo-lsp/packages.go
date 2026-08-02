package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// loadPackage loads the type-checked Go package containing the file at
// path, using overlay (filesystem path -> current buffer content) so that
// unsaved edits are reflected rather than what's on disk.
func loadPackage(ctx context.Context, path string, overlay map[string][]byte) (*packages.Package, *token.FileSet, error) {
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Context: ctx,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedDeps | packages.NeedImports,
		Dir:     filepath.Dir(path),
		Overlay: overlay,
		Fset:    fset,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			return parser.ParseFile(fset, filename, src, parser.ParseComments)
		},
	}

	pkgs, err := packages.Load(cfg, "file="+path)
	if err != nil {
		return nil, nil, err
	}
	if len(pkgs) == 0 {
		return nil, nil, fmt.Errorf("no package found for %s", path)
	}
	if len(pkgs[0].Errors) > 0 && len(pkgs[0].Syntax) == 0 {
		// Couldn't even parse — nothing useful to return.
		return pkgs[0], fset, fmt.Errorf("package failed to load: %w", pkgs[0].Errors[0])
	}

	return pkgs[0], fset, nil
}

// astFileFor finds the *ast.File within pkg.Syntax whose filename (resolved
// via fset) matches path.
func astFileFor(fset *token.FileSet, pkg *packages.Package, path string) *ast.File {
	want := filepath.Clean(path)
	for _, f := range pkg.Syntax {
		got := filepath.Clean(fset.Position(f.Pos()).Filename)
		if got == want {
			return f
		}
	}

	return nil
}
