package main

import (
	"sync"

	"github.com/sazardev/fugo/lsp"
)

// lspState bundles everything the semantic-feature handlers need: the
// server (so they can send notifications like publishDiagnostics), the
// open-document overlay, and the last-published-diagnostics bookkeeping.
type lspState struct {
	srv        *lsp.Server
	docs       *documentStore
	diagsStore *diagnosticsStore

	publishedMu sync.Mutex
	published   map[string][]string // triggering file path -> URIs published last time
}

func newLSPState(srv *lsp.Server) *lspState {
	return &lspState{
		srv:        srv,
		docs:       newDocumentStore(),
		diagsStore: newDiagnosticsStore(),
		published:  make(map[string][]string),
	}
}

func (s *lspState) diagsPublished(path string) []string {
	s.publishedMu.Lock()
	defer s.publishedMu.Unlock()

	return s.published[path]
}

func (s *lspState) setDiagsPublished(path string, uris []string) {
	s.publishedMu.Lock()
	defer s.publishedMu.Unlock()
	s.published[path] = uris
}
