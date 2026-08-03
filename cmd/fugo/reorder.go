package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// reorderProject applies Fugo's opinionated declaration ordering to every
// non-generated, non-test .go file under root, returning how many files it
// actually changed.
func reorderProject(root string) (int, error) {
	changed := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // best-effort walk: skip unreadable paths
		}
		if d.IsDir() {
			if isWatchExcluded(path) {
				return filepath.SkipDir
			}

			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		did, rerr := reorderFile(path)
		if rerr != nil {
			return fmt.Errorf("%s: %w", path, rerr)
		}
		if did {
			changed++
		}

		return nil
	})

	return changed, err
}

// reorderFile rewrites a single file in place, if needed, and reports whether
// it changed. It never touches generated files. The two structural passes
// (function grouping, Router map sorting) each re-parse the file against the
// current line contents before computing their edits — a pass must never
// index into lines using positions computed against a version of the file
// that a previous pass already mutated, or it would splice the wrong text.
func reorderFile(path string) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if bytes.Contains(src, []byte("Code generated")) {
		return false, nil
	}

	lines := splitLinesKeepEnds(src)
	changedAny := false

	funcChanged, lines, err := runPass(path, lines, applyFuncGroupOrder)
	if err != nil {
		return false, err
	}
	changedAny = changedAny || funcChanged

	mapChanged, lines, err := runPass(path, lines, applyRouterMapOrder)
	if err != nil {
		return false, err
	}
	changedAny = changedAny || mapChanged

	if !changedAny {
		return false, nil
	}

	out := strings.Join(lines, "")
	if out == string(src) {
		return false, nil
	}

	return true, os.WriteFile(path, []byte(out), 0o644)
}

// pass computes edits for one file and returns the (possibly resized) line
// slice plus whether anything changed.
type pass func(fset *token.FileSet, f *ast.File, lines []string) ([]string, bool)

// runPass re-parses the current line contents fresh (so the AST positions it
// hands to p always match the lines slice p reads) before applying p.
func runPass(path string, lines []string, p pass) (bool, []string, error) {
	src := []byte(strings.Join(lines, ""))

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return false, lines, err
	}

	next, changed := p(fset, f, lines)

	return changed, next, nil
}

// funcGroupKey buckets a top-level function for reordering: entrypoints
// (main/Build) first, then other exported funcs, then unexported ones. A
// stable sort by this key preserves relative order within each bucket.
func funcGroupKey(fn *ast.FuncDecl) int {
	switch {
	case fn.Name.Name == "main" || fn.Name.Name == "Build":
		return 0
	case fn.Name.IsExported():
		return 1
	default:
		return 2
	}
}

// applyFuncGroupOrder reorders each maximal contiguous run of top-level
// FuncDecls (methods included) within a file by funcGroupKey. Each function's
// own lines (its doc comment, if any, through its closing brace) move as one
// block; the original blank-line spacing between functions is discarded and
// replaced with exactly one blank line between consecutive functions in the
// new order, so the result is well-formed independent of how the input was
// spaced (gofumpt normalizes it further afterward regardless). Runs are
// processed bottom-up so an earlier splice never invalidates the
// still-to-process line numbers of a later run.
func applyFuncGroupOrder(fset *token.FileSet, f *ast.File, lines []string) ([]string, bool) {
	runs := contiguousFuncRuns(f.Decls)
	changed := false

	for _, run := range slices.Backward(runs) {
		if len(run) < 2 {
			continue
		}

		starts := make([]int, len(run))
		ends := make([]int, len(run))
		keys := make([]int, len(run))
		for j, fn := range run {
			starts[j] = funcStartLine(fset, fn)
			ends[j] = fset.Position(fn.End()).Line
			keys[j] = funcGroupKey(fn)
		}

		order := stableOrderByIntKey(keys)
		if isIdentityOrder(order) {
			continue
		}

		groups := make([][]string, len(run))
		for j := range run {
			groups[j] = lines[starts[j]-1 : ends[j]]
		}

		var newSpan []string
		for k, idx := range order {
			if k > 0 {
				newSpan = append(newSpan, "\n")
			}
			newSpan = append(newSpan, groups[idx]...)
		}

		lines = spliceReplace(lines, starts[0], ends[len(run)-1]+1, newSpan)
		changed = true
	}

	return lines, changed
}

// contiguousFuncRuns groups consecutive *ast.FuncDecl entries in decls into
// runs, breaking the run at any other declaration kind.
func contiguousFuncRuns(decls []ast.Decl) [][]*ast.FuncDecl {
	var (
		runs    [][]*ast.FuncDecl
		current []*ast.FuncDecl
	)

	flush := func() {
		if len(current) > 0 {
			runs = append(runs, current)
			current = nil
		}
	}

	for _, d := range decls {
		if fn, ok := d.(*ast.FuncDecl); ok {
			current = append(current, fn)

			continue
		}

		flush()
	}
	flush()

	return runs
}

func funcStartLine(fset *token.FileSet, fn *ast.FuncDecl) int {
	if fn.Doc != nil {
		return fset.Position(fn.Doc.Pos()).Line
	}

	return fset.Position(fn.Pos()).Line
}

// routerLit is one fg.Router(...) map literal found in a file, along with the
// per-entry route keys and their (comment-inclusive) start lines.
type routerLit struct {
	spanStart, spanEnd int // 1-indexed, spanEnd exclusive
	starts             []int
	keys               []string
}

// applyRouterMapOrder sorts the string-keyed entries of any map literal
// passed as the first argument to a call to a function/method named "Router"
// (i.e. fg.Router(map[string]func() fg.Widget{...}, ...)) alphabetically by
// route path. It skips any literal containing a non-string-literal key (e.g.
// a computed key) rather than risk reordering something it can't reason
// about. A CommentMap associates any comment line immediately preceding an
// entry with that entry, so the comment travels with it. Multiple literals in
// one file are applied bottom-up so an earlier splice never invalidates the
// still-to-process line numbers of one higher up.
func applyRouterMapOrder(fset *token.FileSet, f *ast.File, lines []string) ([]string, bool) {
	cmap := ast.NewCommentMap(fset, f, f.Comments)

	var lits []routerLit

	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		if calleeName(call.Fun) != "Router" {
			return true
		}

		lit, ok := call.Args[0].(*ast.CompositeLit)
		if !ok || len(lit.Elts) < 2 {
			return true
		}

		keys := make([]string, len(lit.Elts))
		starts := make([]int, len(lit.Elts))
		for i, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				return true // not a plain key: value map literal — leave it alone
			}

			key, ok := stringLitValue(kv.Key)
			if !ok {
				return true // computed or non-string key — leave it alone
			}

			keys[i] = key
			starts[i] = entryStartLine(fset, cmap, kv)
		}

		lits = append(lits, routerLit{
			spanStart: starts[0],
			spanEnd:   fset.Position(lit.Rbrace).Line, // exclusive: entries end before the closing brace's line
			starts:    starts,
			keys:      keys,
		})

		return true
	})

	changed := false
	for _, lit := range slices.Backward(lits) {
		order := stableOrderByStringKey(lit.keys)
		if isIdentityOrder(order) {
			continue
		}

		newSpan := spliceLineGroups(lines, lit.spanStart, lit.spanEnd, lit.starts, order)
		lines = spliceReplace(lines, lit.spanStart, lit.spanEnd, newSpan)
		changed = true
	}

	return lines, changed
}

// entryStartLine is a map entry's own line, pulled earlier if a CommentMap
// association places a comment immediately before it.
func entryStartLine(fset *token.FileSet, cmap ast.CommentMap, kv *ast.KeyValueExpr) int {
	line := fset.Position(kv.Pos()).Line
	for _, cg := range cmap[kv] {
		if l := fset.Position(cg.Pos()).Line; l < line {
			line = l
		}
	}

	return line
}

// calleeName returns the identifier a call expression resolves to, ignoring
// any package/receiver qualifier ("fg.Router" -> "Router").
func calleeName(fun ast.Expr) string {
	switch e := fun.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	default:
		return ""
	}
}

func stringLitValue(e ast.Expr) (string, bool) {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	v, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	return v, true
}

// stableOrderByIntKey returns, for each output slot, the index of the input
// element that should occupy it, sorted ascending by key and stable on ties.
func stableOrderByIntKey(keys []int) []int {
	order := identityOrder(len(keys))
	sort.SliceStable(order, func(i, j int) bool { return keys[order[i]] < keys[order[j]] })

	return order
}

func stableOrderByStringKey(keys []string) []int {
	order := identityOrder(len(keys))
	sort.SliceStable(order, func(i, j int) bool { return keys[order[i]] < keys[order[j]] })

	return order
}

func identityOrder(n int) []int {
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}

	return order
}

func isIdentityOrder(order []int) bool {
	for i, v := range order {
		if v != i {
			return false
		}
	}

	return true
}

// spliceLineGroups splits lines[spanStart:spanEnd) (1-indexed, exclusive end)
// into contiguous groups starting at each of starts (ascending, starts[0] ==
// spanStart), and returns their concatenation reordered per order (order[i]
// is the index of the group that should occupy output slot i). The result is
// always exactly spanEnd-spanStart lines long.
func spliceLineGroups(lines []string, spanStart, spanEnd int, starts, order []int) []string {
	groups := make([][]string, len(starts))
	for i, s := range starts {
		e := spanEnd
		if i+1 < len(starts) {
			e = starts[i+1]
		}

		groups[i] = lines[s-1 : e-1]
	}

	result := make([]string, 0, spanEnd-spanStart)
	for _, idx := range order {
		result = append(result, groups[idx]...)
	}

	return result
}

// spliceReplace returns a new line slice with lines[spanStart:spanEnd)
// (1-indexed, exclusive end) replaced by newSpan, which may be a different
// length.
func spliceReplace(lines []string, spanStart, spanEnd int, newSpan []string) []string {
	out := make([]string, 0, len(lines)-(spanEnd-spanStart)+len(newSpan))
	out = append(out, lines[:spanStart-1]...)
	out = append(out, newSpan...)
	out = append(out, lines[spanEnd-1:]...)

	return out
}

// splitLinesKeepEnds splits src into lines, each retaining its trailing
// newline (the final line keeps whatever it has, including none), so
// re-joining with "" reproduces the original byte-for-byte when untouched.
func splitLinesKeepEnds(src []byte) []string {
	var lines []string

	start := 0
	for i, b := range src {
		if b == '\n' {
			lines = append(lines, string(src[start:i+1]))
			start = i + 1
		}
	}
	if start < len(src) {
		lines = append(lines, string(src[start:]))
	}

	return lines
}
