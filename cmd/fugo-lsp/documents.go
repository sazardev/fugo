package main

import (
	"sync"

	"github.com/sazardev/fugo/lsp"
)

// documentStore is a thread-safe in-memory map of open document URIs to
// their current buffer content. This is deliberately not the on-disk
// content — the user may have unsaved edits, and every semantic feature in
// this package must see what the editor sees, not what's on disk.
type documentStore struct {
	mu   sync.RWMutex
	docs map[string]string // uri -> content
}

// newDocumentStore creates an empty documentStore.
func newDocumentStore() *documentStore {
	return &documentStore{docs: make(map[string]string)}
}

// set stores (or replaces) the content for uri.
func (d *documentStore) set(uri, content string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.docs[uri] = content
}

// get returns the current content for uri, if it is open.
func (d *documentStore) get(uri string) (string, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	c, ok := d.docs[uri]

	return c, ok
}

// delete removes uri from the store (called on didClose).
func (d *documentStore) delete(uri string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.docs, uri)
}

// overlay returns all open documents as a packages.Config.Overlay-compatible
// map: filesystem paths (not URIs) to their current buffer content.
func (d *documentStore) overlay() map[string][]byte {
	d.mu.RLock()
	defer d.mu.RUnlock()

	out := make(map[string][]byte, len(d.docs))
	for uri, content := range d.docs {
		out[lsp.FilePath(uri)] = []byte(content)
	}

	return out
}
