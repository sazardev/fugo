// Command gen-widgets regenerates the fugo CLI's widget catalog from the fg
// package source. It parses fg/*.go with go/parser (no regex), finds every
// exported top-level function whose return type is a pointer to a struct
// named "*...Widget" (a widget constructor, e.g. Button -> *ButtonWidget),
// and writes the committed file:
//
//   - cmd/fugo/widgets_gen.go   the widgetCatalog used by `fugo widgets`
//
// This keeps `fugo widgets` in sync with the real fg API instead of a
// hand-curated (and inevitably stale) list. Run from the repo root:
//
//	go run ./cmd/gen-widgets
package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// fileGroup maps a source file in fg/ to the catalog group its constructors
// belong to. Files not listed fall back to "Other". Kept intentionally small
// — the main goal of this generator is a complete, accurate widget list, not
// a perfect taxonomy.
var fileGroup = map[string]string{
	"layout.go":    "Layout",
	"container.go": "Layout",
	"sizedbox.go":  "Layout",
	"padding.go":   "Layout",

	"textfield.go": "Input",
	"checkbox.go":  "Input",
	"switch.go":    "Input",
	"slider.go":    "Input",
	"radio.go":     "Input",
	"dropdown.go":  "Input",
	"button.go":    "Input",

	"listview.go": "Scrolling & lists",
	"gridview.go": "Scrolling & lists",
	"scroll.go":   "Scrolling & lists",

	"router.go": "Routing",
}

// groupOrder fixes the print order of known groups; anything else (i.e.
// "Other") is appended last.
var groupOrder = []string{"Layout", "Input", "Scrolling & lists", "Routing"}

type widgetInfo struct {
	ctor string
	desc string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen-widgets:", err)
		os.Exit(1)
	}
}

func run() error {
	widgets, skipped, err := collectWidgets("fg")
	if err != nil {
		return err
	}

	if len(widgets) == 0 {
		return fmt.Errorf("no widget constructors found in fg/")
	}

	if err := writeGo(widgets); err != nil {
		return err
	}

	n := 0
	for _, ws := range widgets {
		n += len(ws)
	}

	fmt.Printf("gen-widgets: wrote %d widgets across %d groups to cmd/fugo/widgets_gen.go\n", n, len(widgets))

	if len(skipped) > 0 {
		fmt.Printf("gen-widgets: %d exported fg funcs had no doc comment: %s\n", len(skipped), strings.Join(skipped, ", "))
	}

	return nil
}

// collectWidgets parses every fg/*.go file and groups widget constructors by
// their inferred catalog group, in file order within each group. It also
// returns the names of constructors that were detected but had no doc
// comment, so the caller can report them.
func collectWidgets(dir string) (map[string][]widgetInfo, []string, error) {
	fset := token.NewFileSet()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", dir, err)
	}

	groups := map[string][]widgetInfo{}

	var skipped []string

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		path := filepath.Join(dir, name)

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, fmt.Errorf("parse %s: %w", path, err)
		}

		group := fileGroup[name]
		if group == "" {
			group = "Other"
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}

			if !returnsWidget(fn) {
				continue
			}

			desc := firstSentence(fn.Doc)
			if desc == "" {
				skipped = append(skipped, fn.Name.Name)
			}

			groups[group] = append(groups[group], widgetInfo{ctor: fn.Name.Name, desc: desc})
		}
	}

	sort.Strings(skipped)

	return groups, skipped, nil
}

// returnsWidget reports whether fn has exactly one result, a pointer to a
// named type ending in "Widget" (e.g. *ButtonWidget). This is how every fg
// widget constructor is shaped; it excludes style/theme helpers (Hex, EdgeAll,
// DarkTheme, ...), tree-building funcs (BuildTree, WithKey, Key, ...), and
// non-widget builders like Span (which returns *TextRun).
func returnsWidget(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}

	star, ok := fn.Type.Results.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	ident, ok := star.X.(*ast.Ident)
	if !ok {
		return false
	}

	return strings.HasSuffix(ident.Name, "Widget")
}

// firstSentence extracts the first sentence of a doc comment's cleaned text,
// suitable as a short catalog description. Returns "" if there is no doc.
func firstSentence(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}

	text := strings.TrimSpace(doc.Text())
	if text == "" {
		return ""
	}

	text = strings.Join(strings.Fields(text), " ")

	if i := strings.Index(text, ". "); i != -1 {
		return text[:i+1]
	}

	return text
}

func writeGo(groups map[string][]widgetInfo) error {
	var order []string

	seen := map[string]bool{}

	for _, g := range groupOrder {
		if _, ok := groups[g]; ok {
			order = append(order, g)
			seen[g] = true
		}
	}

	var rest []string
	for g := range groups {
		if !seen[g] {
			rest = append(rest, g)
		}
	}

	sort.Strings(rest)
	order = append(order, rest...)

	var b strings.Builder

	b.WriteString("// Code generated by cmd/gen-widgets; DO NOT EDIT.\n\n")
	b.WriteString("package main\n\n")
	b.WriteString("// widgetCatalog is a generated view of the fg widget API — every exported\n")
	b.WriteString("// constructor in the fg package whose return type is a *...Widget. Grouping\n")
	b.WriteString("// is inferred from the source file each constructor lives in; see cmd/gen-widgets.\n")
	b.WriteString("var widgetCatalog = []widgetGroup{\n")

	for _, g := range order {
		items := groups[g]
		sort.Slice(items, func(i, j int) bool { return items[i].ctor < items[j].ctor })

		fmt.Fprintf(&b, "\t{%q, []widgetInfo{\n", g)

		for _, w := range items {
			fmt.Fprintf(&b, "\t\t{%q, %q},\n", w.ctor, w.desc)
		}

		b.WriteString("\t}},\n")
	}

	b.WriteString("}\n")

	formatted, err := format.Source([]byte(b.String()))
	if err != nil {
		return fmt.Errorf("format generated Go: %w", err)
	}

	return os.WriteFile(filepath.Join("cmd", "fugo", "widgets_gen.go"), formatted, 0o644)
}
