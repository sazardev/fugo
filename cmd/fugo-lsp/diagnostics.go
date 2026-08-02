package main

import (
	"context"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/sazardev/fugo/fugovet"
	"github.com/sazardev/fugo/lsp"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

// diagnosticsStore keeps the last analysis.Diagnostic set (with its
// SuggestedFixes) per URI, so textDocument/codeAction can offer fixes
// without re-running the analyzers.
type diagnosticsStore struct {
	mu    sync.RWMutex
	byURI map[string][]analysis.Diagnostic
	fset  map[string]*token.FileSet // fset used to produce the diagnostics for that URI
}

func newDiagnosticsStore() *diagnosticsStore {
	return &diagnosticsStore{
		byURI: make(map[string][]analysis.Diagnostic),
		fset:  make(map[string]*token.FileSet),
	}
}

func (d *diagnosticsStore) set(uri string, diags []analysis.Diagnostic, fset *token.FileSet) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.byURI[uri] = diags
	d.fset[uri] = fset
}

func (d *diagnosticsStore) get(uri string) ([]analysis.Diagnostic, *token.FileSet) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.byURI[uri], d.fset[uri]
}

// posRe matches the "file:line:col" format used by packages.Error.Pos.
var posRe = regexp.MustCompile(`^(.+):(\d+):(\d+)$`)

// publishDiagnostics runs a full type-check + fugovet pass for the package
// containing path and publishes textDocument/publishDiagnostics for every
// file it touches (including files that now have zero diagnostics, so the
// client clears stale markers). Callers run this in a detached goroutine (the
// analysis can outlive the didOpen/didSave notification that triggered it),
// so ctx is typically context.Background() from the call site rather than
// the notification's own (soon-cancelled) context.
func (s *lspState) publishDiagnostics(ctx context.Context, path string) {
	pkg, fset, loadErr := loadPackage(ctx, path, s.docs.overlay())

	byFile := make(map[string][]lsp.Diagnostic)
	var analysisDiags []analysis.Diagnostic

	if pkg != nil {
		for _, e := range pkg.Errors {
			file, rng := parsePackagesErrorPos(e)
			if file == "" {
				// No usable position at all; skip rather than attach it
				// to an arbitrary file.
				continue
			}
			byFile[file] = append(byFile[file], lsp.Diagnostic{
				Range:    rng,
				Severity: lsp.SeverityError,
				Message:  e.Msg,
				Source:   "go",
			})
		}

		for _, a := range fugovet.Analyzers {
			diags, err := runAnalyzer(fset, pkg, a)
			if err != nil {
				continue
			}
			for _, d := range diags {
				analysisDiags = append(analysisDiags, d)

				filename := fset.Position(d.Pos).Filename
				content := s.contentFor(filename)
				byFile[filename] = append(byFile[filename], lsp.Diagnostic{
					Range:    tokenPosToLSPRange(fset, content, d.Pos, d.End),
					Severity: lsp.SeverityWarning,
					Message:  d.Message,
					Source:   "fugovet",
				})
			}
		}
	} else if loadErr != nil {
		// Package couldn't be loaded at all (e.g. bad syntax everywhere) —
		// nothing to publish beyond clearing old diagnostics for this file.
		byFile[path] = nil
	}

	// Publish per-file, including files that previously had diagnostics but
	// now have none, so the client clears stale markers.
	prevURIs := s.diagsPublished(path)
	seen := make(map[string]bool)

	for file, diags := range byFile {
		uri := lsp.FileURI(file)
		seen[uri] = true
		if diags == nil {
			diags = []lsp.Diagnostic{}
		}
		_ = s.srv.Notify("textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		})
		s.diagsStore.set(uri, filterDiagsForURI(analysisDiags, fset, file), fset)
	}

	for _, uri := range prevURIs {
		if !seen[uri] {
			_ = s.srv.Notify("textDocument/publishDiagnostics", lsp.PublishDiagnosticsParams{
				URI:         uri,
				Diagnostics: []lsp.Diagnostic{},
			})
			s.diagsStore.set(uri, nil, fset)
		}
	}
	s.setDiagsPublished(path, keys(seen))
}

func filterDiagsForURI(diags []analysis.Diagnostic, fset *token.FileSet, filename string) []analysis.Diagnostic {
	var out []analysis.Diagnostic
	for _, d := range diags {
		if fset.Position(d.Pos).Filename == filename {
			out = append(out, d)
		}
	}

	return out
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	return out
}

// contentFor returns the current buffer content for filename if it's open,
// otherwise reads it from disk.
func (s *lspState) contentFor(filename string) string {
	if c, ok := s.docs.get(lsp.FileURI(filename)); ok {
		return c
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}

	return string(raw)
}

// parsePackagesErrorPos parses a packages.Error's Pos field ("file:line:col")
// into a filename and an lsp.Range. If parsing fails, falls back to line 0
// of an empty filename check — the caller skips diagnostics with an empty
// filename.
func parsePackagesErrorPos(e packages.Error) (string, lsp.Range) {
	m := posRe.FindStringSubmatch(e.Pos)
	if m == nil {
		return "", lsp.Range{}
	}
	file := m[1]
	line, _ := strconv.Atoi(m[2])
	col, _ := strconv.Atoi(m[3])
	if line <= 0 {
		return file, lsp.Range{}
	}

	pos := lsp.Position{Line: uint32(line - 1), Character: uint32(maxInt(col-1, 0))}

	return file, lsp.Range{Start: pos, End: pos}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}

// runAnalyzer runs a single go/analysis.Analyzer against an already-loaded
// packages.Package, wiring up its (shallow) Requires dependency chain
// manually — fugovet's analyzers only require inspect.Analyzer, so this
// does not need to handle deeply nested requirement graphs.
func runAnalyzer(fset *token.FileSet, pkg *packages.Package, a *analysis.Analyzer) ([]analysis.Diagnostic, error) {
	resultOf := map[*analysis.Analyzer]any{}
	for _, req := range a.Requires {
		reqPass := &analysis.Pass{
			Analyzer:   req,
			Fset:       fset,
			Files:      pkg.Syntax,
			Pkg:        pkg.Types,
			TypesInfo:  pkg.TypesInfo,
			TypesSizes: pkg.TypesSizes,
			ResultOf:   map[*analysis.Analyzer]any{},
			Report:     func(analysis.Diagnostic) {},
		}
		res, err := req.Run(reqPass)
		if err != nil {
			return nil, err
		}
		resultOf[req] = res
	}

	var diags []analysis.Diagnostic
	pass := &analysis.Pass{
		Analyzer:   a,
		Fset:       fset,
		Files:      pkg.Syntax,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
		ResultOf:   resultOf,
		Report:     func(d analysis.Diagnostic) { diags = append(diags, d) },
	}
	if _, err := a.Run(pass); err != nil {
		return nil, err
	}

	return diags, nil
}
