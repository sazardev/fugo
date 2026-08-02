package fugovet

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const handlerMissingDoc = `check that fg button constructors have an OnClick handler

fg.Button/FilledButton/FilledTonalButton/OutlinedButton/TextButton/
ElevatedButton/IconButton create a widget that does nothing when clicked
unless .OnClick(...) is registered on it. This analyzer looks for OnClick
either chained directly onto the constructor call, or called later in the
same block on the variable the button was assigned to; if neither is found
it reports the constructor call.

The check is conservative about the "assigned to a variable" case: it only
suppresses the diagnostic when it can find the variable name and confirm
OnClick is called on it somewhere in the enclosing block.`

// HandlerMissing flags fg button constructors with no reachable OnClick.
var HandlerMissing = &analysis.Analyzer{
	Name:     "handlermissing",
	Doc:      handlerMissingDoc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runHandlerMissing,
}

const fgPkgPath = "github.com/sazardev/fugo/fg"

var buttonConstructors = map[string]bool{
	"Button":            true,
	"FilledButton":      true,
	"FilledTonalButton": true,
	"OutlinedButton":    true,
	"TextButton":        true,
	"ElevatedButton":    true,
	"IconButton":        true,
}

func runHandlerMissing(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}
	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		call := n.(*ast.CallExpr)
		if !isButtonConstructor(pass, call) {
			return true
		}

		outer, outerIdx, hasOnClick := climbChainForOnClick(stack, call)
		if hasOnClick {
			return true
		}

		varName, blockIdx := assignedVar(stack, outerIdx, outer)
		if varName != "" {
			if blockIdx >= 0 && blockContainsOnClick(stack[blockIdx], varName) {
				return true
			}
			if blockIdx < 0 {
				// Couldn't locate the enclosing block to check for a later
				// OnClick call — be conservative and stay silent.
				return true
			}
		}

		sel := call.Fun.(*ast.SelectorExpr)
		pass.Reportf(call.Pos(), "%s has no OnClick handler — it will do nothing when clicked", sel.Sel.Name)

		return true
	})

	return nil, nil
}

func isButtonConstructor(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !buttonConstructors[sel.Sel.Name] {
		return false
	}
	fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok || fn.Pkg() == nil {
		return false
	}

	return fn.Pkg().Path() == fgPkgPath
}

// climbChainForOnClick walks up a method-call chain built on top of the
// button constructor call (e.g. fg.Button("x").BgColor(c).OnClick(h)),
// returning the outermost call in that chain and whether OnClick appears
// anywhere along it. outerIdx is the stack index of the outermost call.
func climbChainForOnClick(stack []ast.Node, call *ast.CallExpr) (outer ast.Expr, outerIdx int, hasOnClick bool) {
	cur := ast.Expr(call)
	idx := len(stack) - 1 // index of `call` itself

	for idx-2 >= 0 {
		sel, ok := stack[idx-1].(*ast.SelectorExpr)
		if !ok || sel.X != cur {
			break
		}
		nextCall, ok := stack[idx-2].(*ast.CallExpr)
		if !ok || nextCall.Fun != sel {
			break
		}
		if sel.Sel.Name == "OnClick" {
			hasOnClick = true
		}
		cur = nextCall
		idx -= 2
	}

	return cur, idx, hasOnClick
}

// assignedVar reports the variable name the (possibly chained) button
// expression at stack[outerIdx] is assigned to, and the stack index of the
// nearest enclosing *ast.BlockStmt to search for a later OnClick call. Returns
// ("", -1) if the expression isn't a direct assignment/declaration RHS.
func assignedVar(stack []ast.Node, outerIdx int, outer ast.Expr) (name string, blockIdx int) {
	if outerIdx-1 < 0 {
		return "", -1
	}

	switch p := stack[outerIdx-1].(type) {
	case *ast.AssignStmt:
		for i, rhs := range p.Rhs {
			if rhs == outer && i < len(p.Lhs) {
				if id, ok := p.Lhs[i].(*ast.Ident); ok {
					name = id.Name
				}
			}
		}
	case *ast.ValueSpec:
		for i, v := range p.Values {
			if v == outer && i < len(p.Names) {
				name = p.Names[i].Name
			}
		}
	default:
		return "", -1
	}

	if name == "" {
		return "", -1
	}

	for i := outerIdx - 1; i >= 0; i-- {
		if _, ok := stack[i].(*ast.BlockStmt); ok {
			return name, i
		}
	}

	return name, -1
}

// blockContainsOnClick reports whether block contains a call
// varName.OnClick(...) anywhere in its subtree.
func blockContainsOnClick(block ast.Node, varName string) bool {
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "OnClick" {
			return true
		}
		if id, ok := sel.X.(*ast.Ident); ok && id.Name == varName {
			found = true
		}

		return true
	})

	return found
}
