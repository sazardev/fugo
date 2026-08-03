package fugovet

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const uiImportDoc = `check that a "/ui" package import has an exported Build function

Mirrors part of "fugo doctor"'s project-coherence check: a scaffolded Fugo
app's main.go imports a "<module>/ui" package and calls ui.Build as the
render root. This is a lighter, package-scoped version suitable for
go vet: whenever the analyzed package imports something whose import path
ends in "/ui", it checks that the imported package declares an exported
Build symbol and that it is a function. It does not otherwise validate
go.mod/module coherence — 'fugo doctor' remains the authority for that.`

// UIImport flags a "/ui" import with no exported Build function.
var UIImport = &analysis.Analyzer{
	Name: "uiimport",
	Doc:  uiImportDoc,
	Run:  runUIImport,
}

func runUIImport(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, imp := range file.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasSuffix(path, "/ui") {
				continue
			}

			pkg := importedPackage(pass.Pkg, path)
			if pkg == nil {
				continue // not resolvable from here — stay silent.
			}

			obj := pkg.Scope().Lookup("Build")
			switch obj {
			case nil:
				pass.Reportf(imp.Pos(), "package %q is imported as a UI root but has no exported Build function", path)
			default:
				if _, ok := obj.(*types.Func); !ok {
					pass.Reportf(imp.Pos(), "package %q has a Build symbol but it is not a function", path)
				}
			}
		}
	}

	return nil, nil
}

// importedPackage returns the *types.Package among pkg's imports whose path
// matches, or nil if not found.
func importedPackage(pkg *types.Package, path string) *types.Package {
	for _, imp := range pkg.Imports() {
		if imp.Path() == path {
			return imp
		}
	}

	return nil
}
