package main

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// resolveIdentAt finds the *ast.Ident enclosing byteOffset in file and, if
// found, the types.Object it resolves to (via Uses, falling back to Defs
// for the declaring identifier itself). Returns nil, nil if there is no
// identifier at that position or it doesn't resolve to anything (e.g. it's
// a package name, or the file has type errors around it).
func resolveIdentAt(pkg *packages.Package, fset *token.FileSet, file *ast.File, byteOffset int) (*ast.Ident, types.Object) {
	if file == nil {
		return nil, nil
	}

	pos := file.Pos() + token.Pos(byteOffset)

	path, _ := astutil.PathEnclosingInterval(file, pos, pos)
	for _, n := range path {
		ident, ok := n.(*ast.Ident)
		if !ok {
			continue
		}

		if obj := pkg.TypesInfo.Uses[ident]; obj != nil {
			return ident, obj
		}
		if obj := pkg.TypesInfo.Defs[ident]; obj != nil {
			return ident, obj
		}

		return ident, nil
	}

	return nil, nil
}

// funcDeclFor searches pkg's syntax trees for the *ast.FuncDecl that
// declares fn (matching by name and, for methods, receiver type). Returns
// nil if not found (e.g. fn is declared in a dependency package whose
// syntax wasn't loaded, or fn isn't a function at all).
func funcDeclFor(pkg *packages.Package, fn *types.Func) *ast.FuncDecl {
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name.Name != fn.Name() {
				continue
			}
			if obj := pkg.TypesInfo.Defs[fd.Name]; obj == fn {
				return fd
			}
		}
	}

	return nil
}

// genDeclDocFor searches pkg's syntax trees for the *ast.GenDecl (var/const/
// type) that declares the identifier matching obj, returning its doc
// comment text if found. Handles both a doc comment on the enclosing
// GenDecl and one directly on the ValueSpec/TypeSpec.
//
//nolint:gocognit // walks a genuinely branchy AST shape (GenDecl > {Value,Type}Spec, doc on either); splitting it up would just move the branching into more functions
func genDeclDocFor(pkg *packages.Package, obj types.Object) string {
	for _, f := range pkg.Syntax {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				switch s := spec.(type) {
				case *ast.ValueSpec:
					for _, name := range s.Names {
						if pkg.TypesInfo.Defs[name] == obj {
							if s.Doc != nil {
								return s.Doc.Text()
							}

							return gd.Doc.Text()
						}
					}
				case *ast.TypeSpec:
					if pkg.TypesInfo.Defs[s.Name] == obj {
						if s.Doc != nil {
							return s.Doc.Text()
						}

						return gd.Doc.Text()
					}
				}
			}
		}
	}

	return ""
}

// docFor returns a best-effort doc comment for obj, or "" if none could be
// found. It only handles the common cases (top-level funcs, and var/const/
// type declarations) — anything else is left undocumented rather than risk
// a wrong/misleading comment.
func docFor(pkg *packages.Package, obj types.Object) string {
	switch o := obj.(type) {
	case *types.Func:
		if fd := funcDeclFor(pkg, o); fd != nil && fd.Doc != nil {
			return fd.Doc.Text()
		}
	case *types.Var, *types.Const, *types.TypeName:
		return genDeclDocFor(pkg, obj)
	}

	return ""
}
