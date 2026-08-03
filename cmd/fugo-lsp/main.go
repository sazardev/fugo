// Command fugo-lsp is a standalone Language Server Protocol server for
// Fugo. This binary wires up only the protocol lifecycle (initialize /
// initialized / shutdown / exit) plus no-op document-sync placeholders on
// top of the reusable lsp package; semantic features (hover, completion,
// definition, diagnostics) are added by handlers registered elsewhere.
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/sazardev/fugo/lsp"
)

func main() {
	// stdout is the JSON-RPC channel; any logging must go to stderr (or a
	// file) or it will corrupt the LSP framing.
	logger := log.New(os.Stderr, "fugo-lsp: ", log.LstdFlags)

	srv := lsp.NewServer()
	state := newLSPState(srv)

	srv.HandleRequest("initialize", func(_ context.Context, params json.RawMessage) (any, error) {
		var p lsp.InitializeParams
		if err := json.Unmarshal(params, &p); err != nil {
			logger.Printf("initialize: failed to parse params: %v", err)
		}
		logger.Printf("initialize: rootUri=%q processId=%d", p.RootURI, p.ProcessID)

		return lsp.InitializeResult{
			Capabilities: lsp.ServerCapabilities{
				TextDocumentSync:   lsp.TextDocumentSyncKindFull,
				HoverProvider:      true,
				DefinitionProvider: true,
				CompletionProvider: &lsp.CompletionOptions{
					TriggerCharacters: []string{"."},
				},
				CodeActionProvider:         true,
				DocumentFormattingProvider: true,
			},
		}, nil
	})

	srv.HandleNotification("initialized", func(_ context.Context, _ json.RawMessage) {
		logger.Println("initialized")
	})

	srv.HandleRequest("shutdown", func(_ context.Context, _ json.RawMessage) (any, error) {
		logger.Println("shutdown")

		return nil, nil
	})

	srv.HandleNotification("exit", func(_ context.Context, _ json.RawMessage) {
		logger.Println("exit")
		os.Exit(0)
	})

	srv.HandleNotification("textDocument/didOpen", func(ctx context.Context, params json.RawMessage) {
		var p lsp.DidOpenTextDocumentParams
		_ = json.Unmarshal(params, &p)
		logger.Printf("didOpen: %s", p.TextDocument.URI)

		state.docs.set(p.TextDocument.URI, p.TextDocument.Text)
		go state.publishDiagnostics(context.WithoutCancel(ctx), lsp.FilePath(p.TextDocument.URI))
	})

	srv.HandleNotification("textDocument/didChange", func(_ context.Context, params json.RawMessage) {
		var p lsp.DidChangeTextDocumentParams
		_ = json.Unmarshal(params, &p)
		logger.Printf("didChange: %s", p.TextDocument.URI)

		if len(p.ContentChanges) > 0 {
			state.docs.set(p.TextDocument.URI, p.ContentChanges[len(p.ContentChanges)-1].Text)
		}
	})

	srv.HandleNotification("textDocument/didSave", func(ctx context.Context, params json.RawMessage) {
		var p lsp.DidSaveTextDocumentParams
		_ = json.Unmarshal(params, &p)
		logger.Printf("didSave: %s", p.TextDocument.URI)

		go state.publishDiagnostics(context.WithoutCancel(ctx), lsp.FilePath(p.TextDocument.URI))
	})

	srv.HandleNotification("textDocument/didClose", func(_ context.Context, params json.RawMessage) {
		var p lsp.DidCloseTextDocumentParams
		_ = json.Unmarshal(params, &p)
		logger.Printf("didClose: %s", p.TextDocument.URI)

		state.docs.delete(p.TextDocument.URI)
	})

	srv.HandleRequest("textDocument/hover", state.handleHover)
	srv.HandleRequest("textDocument/definition", state.handleDefinition)
	srv.HandleRequest("textDocument/completion", state.handleCompletion)
	srv.HandleRequest("textDocument/formatting", state.handleFormatting)
	srv.HandleRequest("textDocument/codeAction", state.handleCodeAction)

	ctx := context.Background()
	if err := srv.Serve(ctx, os.Stdin, os.Stdout); err != nil {
		logger.Printf("serve exited: %v", err)
		os.Exit(1)
	}
}
