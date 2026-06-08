package transport

import (
	"testing"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

	"github.com/sazardev/fugo/engine"
)

type testHandler struct{}

func (t *testHandler) HandleEvent(ev *fugov1.ClientEvent)          {}
func (t *testHandler) SetReconciler(stream engine.RenderStream) {}

func TestResolveNetwork_TCP(t *testing.T) {
	network, addr := resolveNetwork("127.0.0.1:9510")
	if network != "tcp" {
		t.Errorf("network = %s, want tcp", network)
	}
	if addr != "127.0.0.1:9510" {
		t.Errorf("addr = %s, want 127.0.0.1:9510", addr)
	}
}

func TestResolveNetwork_UDS(t *testing.T) {
	network, addr := resolveNetwork("/tmp/fugo.sock")
	if network != "unix" {
		t.Errorf("network = %s, want unix", network)
	}
	if addr != "/tmp/fugo.sock" {
		t.Errorf("addr = %s, want /tmp/fugo.sock", addr)
	}
}

func TestResolveNetwork_UDS_Relative(t *testing.T) {
	network, _ := resolveNetwork("fugo.sock")
	if network != "unix" {
		t.Errorf("network = %s, want unix", network)
	}
}

func TestIsUDS(t *testing.T) {
	if !isUDS("/tmp/fugo.sock") {
		t.Error("isUDS(/tmp/fugo.sock) should be true")
	}
	if isUDS("127.0.0.1:9510") {
		t.Error("isUDS(127.0.0.1:9510) should be false")
	}
}

func TestNewServer(t *testing.T) {
	h := &testHandler{}
	s := NewServer(h)
	if s == nil {
		t.Error("NewServer returned nil")
	}
}
