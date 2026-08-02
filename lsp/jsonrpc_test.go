package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"testing"
)

// didOpenMethod is the notification method name used throughout these tests
// and server_test.go.
const didOpenMethod = "textDocument/didOpen"

func TestWriteReadMessageRoundTrip(t *testing.T) {
	var buf bytes.Buffer

	original := &Message{
		JSONRPC: jsonrpcVersion,
		ID:      json.RawMessage(`1`),
		Method:  "initialize",
		Params:  json.RawMessage(`{"processId":123}`),
	}

	if err := WriteMessage(&buf, original); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}

	// Splice in an extra, unrelated header before the blank-line terminator
	// to exercise that ReadMessage ignores headers other than
	// Content-Length.
	raw := buf.Bytes()
	sep := []byte("\r\n\r\n")
	idx := bytes.Index(raw, sep)
	if idx < 0 {
		t.Fatalf("no header/body separator found in written message")
	}
	var spliced bytes.Buffer
	spliced.Write(raw[:idx])
	spliced.WriteString("\r\nContent-Type: application/vscode-jsonrpc; charset=utf-8")
	spliced.Write(raw[idx:])

	r := bufio.NewReader(&spliced)
	got, err := ReadMessage(r)
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}

	if got.JSONRPC != "2.0" {
		t.Errorf("JSONRPC = %q, want %q", got.JSONRPC, "2.0")
	}
	if got.Method != original.Method {
		t.Errorf("Method = %q, want %q", got.Method, original.Method)
	}
	if string(got.ID) != string(original.ID) {
		t.Errorf("ID = %s, want %s", got.ID, original.ID)
	}
	if string(got.Params) != string(original.Params) {
		t.Errorf("Params = %s, want %s", got.Params, original.Params)
	}
}

func TestReadMessageMultipleMessages(t *testing.T) {
	var buf bytes.Buffer

	msg1 := &Message{JSONRPC: jsonrpcVersion, Method: didOpenMethod, Params: json.RawMessage(`{"a":1}`)}
	msg2 := &Message{JSONRPC: jsonrpcVersion, ID: json.RawMessage(`"abc"`), Method: "shutdown"}

	if err := WriteMessage(&buf, msg1); err != nil {
		t.Fatalf("WriteMessage msg1: %v", err)
	}
	if err := WriteMessage(&buf, msg2); err != nil {
		t.Fatalf("WriteMessage msg2: %v", err)
	}

	r := bufio.NewReader(&buf)

	got1, err := ReadMessage(r)
	if err != nil {
		t.Fatalf("ReadMessage msg1: %v", err)
	}
	if got1.Method != didOpenMethod {
		t.Errorf("msg1 Method = %q", got1.Method)
	}

	got2, err := ReadMessage(r)
	if err != nil {
		t.Fatalf("ReadMessage msg2: %v", err)
	}
	if got2.Method != "shutdown" || string(got2.ID) != `"abc"` {
		t.Errorf("msg2 = %+v", got2)
	}
}

func TestReadMessageMissingContentLength(t *testing.T) {
	r := bufio.NewReader(bytes.NewBufferString("Content-Type: foo\r\n\r\n"))
	if _, err := ReadMessage(r); err == nil {
		t.Fatal("expected error for missing Content-Length header, got nil")
	}
}
