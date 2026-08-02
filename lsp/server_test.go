package lsp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestServeInitializeRoundTrip feeds a synthetic "initialize" request into
// Serve via an in-memory buffer and verifies the response is a well-formed
// JSON-RPC message carrying the same id as the request and the result the
// registered handler returned.
func TestServeInitializeRoundTrip(t *testing.T) {
	srv := NewServer()

	type initResult struct {
		OK bool `json:"ok"`
	}

	var gotParams InitializeParams
	srv.HandleRequest("initialize", func(_ context.Context, params json.RawMessage) (any, error) {
		if err := json.Unmarshal(params, &gotParams); err != nil {
			t.Errorf("unmarshal params: %v", err)
		}

		return initResult{OK: true}, nil
	})

	var in bytes.Buffer
	req := &Message{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`42`),
		Method:  "initialize",
		Params:  json.RawMessage(`{"processId":7,"rootUri":"file:///tmp/proj"}`),
	}
	if err := WriteMessage(&in, req); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	var out bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- srv.Serve(ctx, &in, &out)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after input EOF (timeout)")
	}

	if gotParams.ProcessID != 7 || gotParams.RootURI != "file:///tmp/proj" {
		t.Errorf("handler did not see expected params: %+v", gotParams)
	}

	r := bufio.NewReader(&out)
	resp, err := ReadMessage(r)
	if err != nil {
		t.Fatalf("ReadMessage response: %v", err)
	}

	if string(resp.ID) != string(req.ID) {
		t.Errorf("response ID = %s, want %s", resp.ID, req.ID)
	}
	if resp.Error != nil {
		t.Fatalf("response had unexpected error: %+v", resp.Error)
	}

	var result initResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if !result.OK {
		t.Errorf("result.OK = false, want true")
	}
}

// TestServeMethodNotFound verifies that a request for an unregistered
// method gets a JSON-RPC error response rather than being silently
// dropped or crashing the server.
func TestServeMethodNotFound(t *testing.T) {
	srv := NewServer()

	var in bytes.Buffer
	req := &Message{JSONRPC: "2.0", ID: json.RawMessage(`1`), Method: "textDocument/hover"}
	if err := WriteMessage(&in, req); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	var out bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Serve(ctx, &in, &out); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	resp, err := ReadMessage(bufio.NewReader(&out))
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error response for unknown method, got none")
	}
	if resp.Error.Code != MethodNotFound {
		t.Errorf("error code = %d, want %d", resp.Error.Code, MethodNotFound)
	}
}

// TestServeNotificationNoResponse verifies notifications never produce a
// response message, even when a handler is registered.
func TestServeNotificationNoResponse(t *testing.T) {
	srv := NewServer()

	called := make(chan struct{}, 1)
	srv.HandleNotification("textDocument/didOpen", func(_ context.Context, _ json.RawMessage) {
		called <- struct{}{}
	})

	var in bytes.Buffer
	note := &Message{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: json.RawMessage(`{}`)}
	if err := WriteMessage(&in, note); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	var out bytes.Buffer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := srv.Serve(ctx, &in, &out); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("notification handler was not called")
	}

	if out.Len() != 0 {
		t.Errorf("expected no output for a notification, got %d bytes", out.Len())
	}
}
