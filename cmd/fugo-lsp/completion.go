package main

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/sazardev/fugo/lsp"
	"golang.org/x/tools/go/packages"
)

const fugoWidgetPkg = "github.com/sazardev/fugo/fg"

// staticFgCompletions is the fallback list used when the file can't be
// type-checked at all (e.g. mid-edit syntax error) — just enough that the
// user is never left with zero suggestions after typing "fg.".
var staticFgCompletions = []string{
	"Text", "Button", "Container", "Row", "Column", "Center", "SizedBox", "Router",
}

// handleCompletion implements "textDocument/completion". See the package
// doc comment on completion strategy: (a) fg.-prefixed completion of real,
// type-checked exported functions from the fg package; (b) generic member
// completion (methods/fields) for a resolvable expression before the dot;
// (c) a static fallback list when the package can't be loaded/resolved.
func (s *lspState) handleCompletion(ctx context.Context, params json.RawMessage) (any, error) {
	var p lsp.CompletionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	content, ok := s.docs.get(p.TextDocument.URI)
	if !ok {
		return lsp.CompletionList{}, nil
	}
	path := lsp.FilePath(p.TextDocument.URI)

	prefix, hasDot := selectorPrefix(content, p.Position)
	if !hasDot {
		return lsp.CompletionList{IsIncomplete: false, Items: nil}, nil
	}

	pkg, fset, err := loadPackage(ctx, path, s.docs.overlay())
	if err != nil || pkg == nil {
		return lsp.CompletionList{IsIncomplete: false, Items: staticItems()}, nil
	}

	file := astFileFor(fset, pkg, path)
	if file == nil {
		return lsp.CompletionList{IsIncomplete: false, Items: staticItems()}, nil
	}

	// (a) Is `prefix` the local alias for the fg package?
	if alias, ok := fgImportAlias(file); ok && alias == prefix {
		if items := fgPackageCompletions(pkg); items != nil {
			return lsp.CompletionList{IsIncomplete: false, Items: items}, nil
		}
	}

	// (b) Generic member completion for a resolvable identifier expression.
	if items := memberCompletions(pkg, fset, file, content, p.Position, prefix); items != nil {
		return lsp.CompletionList{IsIncomplete: false, Items: items}, nil
	}

	// (c) Fallback.
	return lsp.CompletionList{IsIncomplete: false, Items: staticItems()}, nil
}

// selectorPrefix looks at the text immediately before the cursor on the
// current line and, if it matches "<ident>.", returns <ident> and true.
func selectorPrefix(content string, pos lsp.Position) (string, bool) {
	lines := splitLines(content)
	if int(pos.Line) >= len(lines) {
		return "", false
	}
	line := lines[pos.Line]

	byteCol := utf16OffsetToByteOffset(line, int(pos.Character))
	if byteCol > len(line) {
		byteCol = len(line)
	}
	before := line[:byteCol]

	if !strings.HasSuffix(before, ".") {
		return "", false
	}
	before = strings.TrimSuffix(before, ".")

	// Walk back over identifier characters.
	i := len(before)
	for i > 0 {
		c := before[i-1]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			i--

			continue
		}

		break
	}
	ident := before[i:]
	if ident == "" {
		return "", false
	}

	return ident, true
}

// fgImportAlias returns the local identifier used in file to refer to the
// fg package (usually "fg", but respects an explicit alias), and whether it
// is imported at all.
func fgImportAlias(file *ast.File) (string, bool) {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if path != fugoWidgetPkg {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name, true
		}

		return "fg", true
	}

	return "", false
}

// fgPackageCompletions enumerates every exported *types.Func in the fg
// package's scope, per the task's explicit instruction to show the full
// catalog (constructors and helpers alike) rather than filter heuristically.
func fgPackageCompletions(pkg *packages.Package) []lsp.CompletionItem {
	fgPkg := pkg.Imports[fugoWidgetPkg]
	if fgPkg == nil || fgPkg.Types == nil {
		return nil
	}

	scope := fgPkg.Types.Scope()
	names := scope.Names()
	items := make([]lsp.CompletionItem, 0, len(names))
	for _, name := range names {
		if !ast.IsExported(name) {
			continue
		}
		obj := scope.Lookup(name)
		fn, ok := obj.(*types.Func)
		if !ok {
			continue
		}
		items = append(items, lsp.CompletionItem{
			Label:      name,
			Kind:       lsp.CompletionItemKindFunction,
			Detail:     types.ObjectString(fn, types.RelativeTo(fgPkg.Types)),
			InsertText: name,
		})
	}

	return items
}

// memberCompletions handles the generic (non-fg) case: `prefix` is a Go
// expression (kept deliberately simple — a bare identifier, not an
// arbitrary chained expression) that resolves to a value; enumerate its
// method set and, for structs, exported fields.
//
// This intentionally does not attempt full chained-selector-expression
// resolution (e.g. `a.b.c.`) — that requires re-parsing the partial
// expression text (which may not even be valid Go while the user is
// mid-edit) and locating it precisely in the AST. Scope: single identifier
// only, to keep this correct and simple rather than broad and fragile.
//
// Prefix filtering is left to the LSP client (the standard division of
// labor — most clients fuzzy-filter completion items themselves), so the
// prefix argument is currently unused here.
func memberCompletions(pkg *packages.Package, fset *token.FileSet, file *ast.File, content string, pos lsp.Position, _ string) []lsp.CompletionItem {
	offset := lspPositionToByteOffset(content, pos)
	// Point somewhere inside the identifier that precedes the dot: the
	// dot sits at offset-1, so offset-2 lands on the identifier's last
	// byte (prefix is always ASCII, see selectorPrefix).
	_, obj := resolveIdentAt(pkg, fset, file, offset-2)
	if obj == nil {
		return nil
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return nil
	}
	typ := v.Type()

	var items []lsp.CompletionItem

	// Method set (works for both value and pointer receivers via the
	// pointer method set).
	ptrType := typ
	if _, isPtr := typ.(*types.Pointer); !isPtr {
		ptrType = types.NewPointer(typ)
	}
	mset := types.NewMethodSet(ptrType)
	for i := range mset.Len() {
		sel := mset.At(i)
		fn, ok := sel.Obj().(*types.Func)
		if !ok || !fn.Exported() {
			continue
		}
		items = append(items, lsp.CompletionItem{
			Label:      fn.Name(),
			Kind:       lsp.CompletionItemKindMethod,
			Detail:     types.ObjectString(fn, types.RelativeTo(pkg.Types)),
			InsertText: fn.Name(),
		})
	}

	// Exported struct fields, if the underlying type is a struct.
	underlying := typ
	if p, ok := underlying.(*types.Pointer); ok {
		underlying = p.Elem()
	}
	if st, ok := underlying.Underlying().(*types.Struct); ok {
		for i := range st.NumFields() {
			f := st.Field(i)
			if !f.Exported() {
				continue
			}
			items = append(items, lsp.CompletionItem{
				Label:      f.Name(),
				Kind:       lsp.CompletionItemKindField,
				Detail:     types.ObjectString(f, types.RelativeTo(pkg.Types)),
				InsertText: f.Name(),
			})
		}
	}

	return items
}

func staticItems() []lsp.CompletionItem {
	items := make([]lsp.CompletionItem, 0, len(staticFgCompletions))
	for _, name := range staticFgCompletions {
		items = append(items, lsp.CompletionItem{
			Label:      name,
			Kind:       lsp.CompletionItemKindSnippet,
			InsertText: name,
		})
	}

	return items
}
