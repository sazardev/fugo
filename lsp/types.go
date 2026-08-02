package lsp

import "encoding/json"

// This file contains only the LSP 3.17 types needed for the lifecycle
// methods (initialize/initialized/shutdown/exit) and the document sync +
// language feature methods listed in the package's design brief. Field
// names follow the LSP spec's camelCase JSON encoding exactly so that real
// clients (VS Code, neovim, ...) can talk to a server built on this
// package. Sync is full-document only (TextDocumentSyncKindFull); there is
// no incremental sync support.

// Position is a zero-based line/character offset (UTF-16 code units, per
// the LSP spec) into a text document.
type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

// Range is a start/end pair of Positions.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location identifies a range within a document identified by URI.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentItem is the full content + metadata of a document as sent by
// textDocument/didOpen.
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int32  `json:"version"`
	Text       string `json:"text"`
}

// TextDocumentIdentifier identifies a document by URI only.
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier identifies a document by URI plus the
// version of the edit being described.
type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int32  `json:"version"`
}

// TextDocumentPositionParams is the common shape shared by hover,
// definition, and completion requests: a document plus a position in it.
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// Diagnostic severity levels, per the LSP spec.
const (
	SeverityError       = 1
	SeverityWarning     = 2
	SeverityInformation = 3
	SeverityHint        = 4
)

// Diagnostic describes a single issue (error/warning/info/hint) at a range
// within a document.
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Message  string `json:"message"`
	Source   string `json:"source,omitempty"`
}

// PublishDiagnosticsParams is sent server -> client via the
// "textDocument/publishDiagnostics" notification.
type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// DidOpenTextDocumentParams is the payload of "textDocument/didOpen".
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentContentChangeEvent describes a change to a document. Since
// this package only supports full sync (TextDocumentSyncKindFull), Text
// always carries the complete new document content.
type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

// DidChangeTextDocumentParams is the payload of "textDocument/didChange"
// under full sync: ContentChanges will contain exactly one element whose
// Text is the entire new document.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidSaveTextDocumentParams is the payload of "textDocument/didSave". Text
// is optional per spec (only present if the server asked for it); included
// here for completeness.
type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Text         string                 `json:"text,omitempty"`
}

// DidCloseTextDocumentParams is the payload of "textDocument/didClose".
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// HoverParams is the payload of "textDocument/hover".
type HoverParams struct {
	TextDocumentPositionParams
}

// MarkupContent is a markdown or plaintext content blob, per the LSP spec.
type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown" or "plaintext"
	Value string `json:"value"`
}

// Hover is the result of "textDocument/hover".
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// DefinitionParams is the payload of "textDocument/definition".
type DefinitionParams struct {
	TextDocumentPositionParams
}

// CompletionParams is the payload of "textDocument/completion".
type CompletionParams struct {
	TextDocumentPositionParams
}

// CompletionItem kinds, per the LSP spec (subset commonly used).
const (
	CompletionItemKindText      = 1
	CompletionItemKindMethod    = 2
	CompletionItemKindFunction  = 3
	CompletionItemKindField     = 5
	CompletionItemKindVariable  = 6
	CompletionItemKindClass     = 7
	CompletionItemKindKeyword   = 14
	CompletionItemKindSnippet   = 15
	CompletionItemKindProperty  = 10
	CompletionItemKindModule    = 9
	CompletionItemKindInterface = 8
	CompletionItemKindConstant  = 21
)

// CompletionItem is a single completion candidate.
type CompletionItem struct {
	Label      string `json:"label"`
	Kind       int    `json:"kind,omitempty"`
	Detail     string `json:"detail,omitempty"`
	InsertText string `json:"insertText,omitempty"`
}

// CompletionList is the result of "textDocument/completion".
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// CodeActionContext carries the diagnostics currently known at the
// requested range, so a code action can offer fixes for them.
type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// CodeActionParams is the payload of "textDocument/codeAction".
type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

// TextEdit is a single replacement of Range's content with NewText.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// WorkspaceEdit maps a document URI to the list of edits to apply to it.
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes,omitempty"`
}

// CodeAction is a single suggested fix/refactor, optionally carrying the
// diagnostics it resolves and the edit to apply.
type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
}

// DocumentFormattingParams is the payload of "textDocument/formatting".
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// InitializeParams is the payload of "initialize". Capabilities is kept as
// raw JSON: this package does not need to interpret client capabilities in
// detail.
type InitializeParams struct {
	ProcessID    int             `json:"processId"`
	RootURI      string          `json:"rootUri"`
	Capabilities json.RawMessage `json:"capabilities"`
}

// TextDocumentSyncKindFull announces full-document sync (the only mode
// this package supports).
const TextDocumentSyncKindFull = 1

// CompletionOptions announces completion support and its trigger
// characters.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// ServerCapabilities announces what the server supports in response to
// "initialize".
type ServerCapabilities struct {
	TextDocumentSync           int                `json:"textDocumentSync"`
	HoverProvider              bool               `json:"hoverProvider,omitempty"`
	DefinitionProvider         bool               `json:"definitionProvider,omitempty"`
	CompletionProvider         *CompletionOptions `json:"completionProvider,omitempty"`
	CodeActionProvider         bool               `json:"codeActionProvider,omitempty"`
	DocumentFormattingProvider bool               `json:"documentFormattingProvider,omitempty"`
}

// InitializeResult is the response to "initialize".
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}
