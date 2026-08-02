package fugovet

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const updateMissingDoc = `check that widget setter calls are followed by ctx.Update()/ctx.UpdateNow()

The Fugo render loop only re-renders when the retained widget tree is
mutated AND the *fugo.Context passed to the enclosing function/handler has
Update() or UpdateNow() called on it. This analyzer flags a "Set*" method
call (e.g. count.SetText(...)) made inside a function that has access to a
*fugo.Context when no call to that context's Update()/UpdateNow() can be
found anywhere in the same function (or closure) body.

The check is conservative: it only fires when it can identify, by static
type, the *fugo.Context parameter feeding the enclosing scope. If it can't
determine that with confidence, it stays silent rather than risk a false
positive.`

// UpdateMissing flags Set* widget mutations that have no reachable
// ctx.Update()/UpdateNow() call in the same function scope.
var UpdateMissing = &analysis.Analyzer{
	Name:     "updatemissing",
	Doc:      updateMissingDoc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runUpdateMissing,
}

const fugoPkgPath = "github.com/sazardev/fugo"

func runUpdateMissing(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isSetterCall(pass, call) {
			return true
		}

		ctxName, innerBody := resolveCtxScope(pass, stack)
		if ctxName == "" || innerBody == nil {
			return true // can't determine the Context scope confidently — stay silent.
		}
		if containsUpdateCall(innerBody, ctxName) {
			return true
		}

		sel := call.Fun.(*ast.SelectorExpr)
		diag := analysis.Diagnostic{
			Pos: call.Pos(),
			Message: fmt.Sprintf(
				"%s call is not followed by %s.Update()/%s.UpdateNow() in this scope — the change won't be rendered",
				sel.Sel.Name, ctxName, ctxName,
			),
		}
		if fix, ok := updateFixFor(stack, call, ctxName); ok {
			diag.SuggestedFixes = []analysis.SuggestedFix{fix}
		}
		pass.Report(diag)

		return true
	})

	return nil, nil
}

// resolveCtxScope walks the ancestor stack of a Set* call (innermost first)
// to find: (1) the nearest enclosing function body — the scope that must
// contain the ctx.Update() call, and (2) the name of the *fugo.Context
// parameter that is in scope there, which may belong to an outer function
// (closures capture it). Returns ("", nil) if no such parameter is found
// anywhere up the stack.
func resolveCtxScope(pass *analysis.Pass, stack []ast.Node) (ctxName string, innerBody *ast.BlockStmt) {
	haveInner := false
	for i := len(stack) - 2; i >= 0; i-- {
		var ft *ast.FuncType
		var body *ast.BlockStmt
		switch fn := stack[i].(type) {
		case *ast.FuncLit:
			ft, body = fn.Type, fn.Body
		case *ast.FuncDecl:
			ft, body = fn.Type, fn.Body
		default:
			continue
		}
		if !haveInner {
			innerBody = body
			haveInner = true
		}
		if name := funcCtxParam(pass, ft); name != "" {
			return name, innerBody
		}
	}

	return "", nil
}

// funcCtxParam returns the name of the first parameter of ft whose type is
// *fugo.Context, or "" if none.
func funcCtxParam(pass *analysis.Pass, ft *ast.FuncType) string {
	if ft == nil || ft.Params == nil {
		return ""
	}
	for _, field := range ft.Params.List {
		if !isFugoContextPtr(pass.TypesInfo.TypeOf(field.Type)) {
			continue
		}
		for _, name := range field.Names {
			if name.Name != "" && name.Name != "_" {
				return name.Name
			}
		}
	}

	return ""
}

func isFugoContextPtr(t types.Type) bool {
	if t == nil {
		return false
	}
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()

	return obj != nil && obj.Name() == "Context" && obj.Pkg() != nil && obj.Pkg().Path() == fugoPkgPath
}

// isSetterCall reports whether call looks like widget.SetFoo(...): a
// selector call whose method name starts with "Set" followed by an
// uppercase letter, AND whose receiver is a value of a type declared in
// package fg (an actual widget) — not, say, a *fugo.WindowController or
// clipboard/file-dialog handle, whose Set* methods (if any) don't
// participate in the retained widget tree at all and so have nothing to do
// with ctx.Update().
func isSetterCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	if !strings.HasPrefix(name, "Set") || len(name) <= len("Set") {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name[len("Set"):])
	if !unicode.IsUpper(r) {
		return false
	}

	return isFgWidgetType(pass.TypesInfo.TypeOf(sel.X))
}

// isFgWidgetType reports whether t is (a pointer to) a named type declared
// in package github.com/sazardev/fugo/fg.
func isFgWidgetType(t types.Type) bool {
	if t == nil {
		return false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()

	return obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == fgPkgPath
}

// containsUpdateCall reports whether body contains a call to
// ctxName.Update() or ctxName.UpdateNow() anywhere in its subtree, including
// nested closures — deliberately unbounded, to avoid false positives when
// the update happens to be issued from a deeper nested handler.
func containsUpdateCall(body ast.Node, ctxName string) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name != "Update" && sel.Sel.Name != "UpdateNow" {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && id.Name == ctxName {
			found = true
		}

		return true
	})

	return found
}

// updateFixFor builds a SuggestedFix that appends "<ctx>.Update()" right
// after the statement containing call, but only for the simplest, safest
// shape: call is the entire expression of a bare expression statement (so
// appending a statement after it cannot change control flow, e.g. it isn't
// nested inside a return or condition).
func updateFixFor(stack []ast.Node, call *ast.CallExpr, ctxName string) (analysis.SuggestedFix, bool) {
	// Find the nearest ancestor statement of call. Only fix the safe case:
	// call is the entire expression of a bare expression statement, so
	// appending a new statement right after it cannot change control flow
	// (unlike e.g. a call nested in a return or an if-condition).
	for i := len(stack) - 2; i >= 0; i-- {
		stmt, ok := stack[i].(ast.Stmt)
		if !ok {
			continue
		}
		es, ok := stmt.(*ast.ExprStmt)
		if !ok || es.X != call {
			return analysis.SuggestedFix{}, false
		}

		return analysis.SuggestedFix{
			Message: fmt.Sprintf("insert %s.Update()", ctxName),
			TextEdits: []analysis.TextEdit{{
				Pos:     es.End(),
				End:     es.End(),
				NewText: []byte("\n\t" + ctxName + ".Update()"),
			}},
		}, true
	}

	return analysis.SuggestedFix{}, false
}
