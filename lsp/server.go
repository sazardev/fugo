package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// RequestHandler processes a request and returns a result (marshaled into
// the response's Result field) or an error.
type RequestHandler func(ctx context.Context, params json.RawMessage) (any, error)

// NotificationHandler processes a notification. There is no response.
type NotificationHandler func(ctx context.Context, params json.RawMessage)

// Server is a JSON-RPC 2.0 / LSP method dispatcher. It owns no transport by
// itself — Serve reads from an io.Reader and writes to an io.Writer, so it
// works equally over stdio or any other stream.
//
// Server is safe for concurrent use: handlers may be registered before
// Serve is called, and Notify may be called concurrently from within
// handler goroutines while Serve is running.
type Server struct {
	mu            sync.RWMutex
	requests      map[string]RequestHandler
	notifications map[string]NotificationHandler

	writeMu sync.Mutex
	w       io.Writer
}

// NewServer creates an empty Server with no handlers registered.
func NewServer() *Server {
	return &Server{
		requests:      make(map[string]RequestHandler),
		notifications: make(map[string]NotificationHandler),
	}
}

// HandleRequest registers a handler for a request-style LSP method (e.g.
// "initialize", "textDocument/hover"). Registering a handler for a method
// that already has one replaces it.
func (s *Server) HandleRequest(method string, h RequestHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests[method] = h
}

// HandleNotification registers a handler for a notification-style LSP
// method (e.g. "textDocument/didOpen"). Registering a handler for a method
// that already has one replaces it.
func (s *Server) HandleNotification(method string, h NotificationHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifications[method] = h
}

// Notify sends a server-to-client notification (e.g.
// "textDocument/publishDiagnostics"). It is safe to call concurrently,
// including from background goroutines spawned by request/notification
// handlers, for as long as Serve is running (or even before/after, as long
// as the writer is valid).
func (s *Server) Notify(method string, params any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	msg := &Message{
		JSONRPC: jsonrpcVersion,
		Method:  method,
		Params:  raw,
	}

	return s.writeMessage(msg)
}

func (s *Server) writeMessage(msg *Message) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.w == nil {
		return errors.New("lsp: server is not serving (no writer set)")
	}

	return WriteMessage(s.w, msg)
}

// Serve runs the main dispatch loop: it reads LSP-framed JSON-RPC messages
// from r, dispatches each to the handler registered for its method (each
// message is handled in its own goroutine so a slow request cannot block
// the read loop or other in-flight requests), and writes responses /
// notifications to w. Writes to w are serialized internally so concurrent
// handlers (and Notify calls) never interleave partial messages.
//
// Serve returns nil when r reaches EOF (the client closed the stream,
// typically after "exit"), or the context error if ctx is canceled first.
// It blocks until all in-flight handler goroutines have finished.
func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	s.writeMu.Lock()
	s.w = w
	s.writeMu.Unlock()

	br := bufio.NewReader(r)

	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		msg, err := ReadMessage(br)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		wg.Add(1)
		go func(msg *Message) {
			defer wg.Done()
			s.dispatch(ctx, msg)
		}(msg)
	}
}

//nolint:nestif // dispatch legitimately branches on request-vs-notification, then found-vs-not-found within each
func (s *Server) dispatch(ctx context.Context, msg *Message) {
	isRequest := len(msg.ID) > 0

	if isRequest {
		s.mu.RLock()
		h, ok := s.requests[msg.Method]
		s.mu.RUnlock()

		if !ok {
			_ = s.writeMessage(&Message{
				JSONRPC: jsonrpcVersion,
				ID:      msg.ID,
				Error: &RPCError{
					Code:    MethodNotFound,
					Message: "method not found: " + msg.Method,
				},
			})

			return
		}

		result, err := h(ctx, msg.Params)
		if err != nil {
			var rpcErr *RPCError
			if !errors.As(err, &rpcErr) {
				rpcErr = &RPCError{Code: InternalError, Message: err.Error()}
			}
			_ = s.writeMessage(&Message{
				JSONRPC: jsonrpcVersion,
				ID:      msg.ID,
				Error:   rpcErr,
			})

			return
		}

		raw, err := json.Marshal(result)
		if err != nil {
			_ = s.writeMessage(&Message{
				JSONRPC: jsonrpcVersion,
				ID:      msg.ID,
				Error: &RPCError{
					Code:    InternalError,
					Message: fmt.Sprintf("failed to marshal result: %v", err),
				},
			})

			return
		}

		_ = s.writeMessage(&Message{
			JSONRPC: jsonrpcVersion,
			ID:      msg.ID,
			Result:  raw,
		})

		return
	}

	// Notification: no response, regardless of whether a handler exists.
	s.mu.RLock()
	h, ok := s.notifications[msg.Method]
	s.mu.RUnlock()

	if ok {
		h(ctx, msg.Params)
	}
}
