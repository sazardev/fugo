// Package lsp implements the protocol core of a from-scratch Language Server
// Protocol (LSP) server: JSON-RPC 2.0 framing over stdio and a method
// dispatcher. It does not implement any Fugo-specific semantic analysis —
// that is layered on top by registering handlers via Server.HandleRequest /
// Server.HandleNotification.
package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Message is a JSON-RPC 2.0 message as used by LSP. Depending on which
// fields are set it represents a request (ID+Method), a notification
// (Method only, no ID), or a response (ID + Result or Error).
type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message)
}

// Standard JSON-RPC / LSP error codes.
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// ReadMessage reads a single LSP-framed JSON-RPC message from r. The LSP
// framing is a set of "Header: value\r\n" lines terminated by a blank line
// ("\r\n"), where the only header this implementation interprets is
// Content-Length; any other header (e.g. Content-Type) is read and ignored.
// The header block is followed by exactly Content-Length bytes of JSON.
//
// ReadMessage is not safe for concurrent use on the same *bufio.Reader.
func ReadMessage(r *bufio.Reader) (*Message, error) {
	var contentLength int
	haveLength := false

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			// Blank line: end of headers.
			break
		}

		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)

		if strings.EqualFold(name, "Content-Length") {
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("lsp: invalid Content-Length %q: %w", value, err)
			}
			contentLength = n
			haveLength = true
		}
		// Any other header (e.g. Content-Type) is ignored.
	}

	if !haveLength {
		return nil, fmt.Errorf("lsp: message missing Content-Length header")
	}

	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, fmt.Errorf("lsp: invalid JSON body: %w", err)
	}

	return &msg, nil
}

// WriteMessage serializes msg to JSON and writes it to w using the LSP
// Content-Length framing.
//
// WriteMessage performs a single Write call with the fully assembled
// header+body buffer, but it does not itself serialize concurrent callers —
// if multiple goroutines may call WriteMessage on the same w concurrently,
// the caller must provide external locking (Server does this internally for
// its own writer).
func WriteMessage(w io.Writer, msg *Message) error {
	if msg.JSONRPC == "" {
		msg.JSONRPC = "2.0"
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Content-Length: %d\r\n\r\n", len(body))
	buf.Write(body)

	_, err = w.Write(buf.Bytes())

	return err
}
