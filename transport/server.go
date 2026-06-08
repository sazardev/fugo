package transport

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

	"github.com/sazardev/fugo/engine"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

// Server implements the FugoRender gRPC service, bridging the bidirectional
// render stream to the application: it pushes render payloads to the Flutter
// client and forwards inbound client events to the AppHandler.
type Server struct {
	fugov1.UnimplementedFugoRenderServer
	app AppHandler
}

// AppHandler is the application-side contract the transport depends on. The
// server forwards each inbound ClientEvent via HandleEvent and, when a client
// connects, hands the outbound stream to the app via SetReconciler.
type AppHandler interface {
	HandleEvent(ev *fugov1.ClientEvent)
	SetReconciler(stream engine.RenderStream)
}

// NewServer returns a Server that routes events to and renders through app.
func NewServer(app AppHandler) *Server {
	return &Server{app: app}
}

// RenderStream handles the bidirectional render stream for one client
// connection. It registers the outbound stream with the app, then blocks
// receiving client events and forwarding them until the client disconnects,
// returning the receive error that ended the stream.
func (s *Server) RenderStream(stream fugov1.FugoRender_RenderStreamServer) error {
	adapter := &grpcStreamAdapter{stream: stream}
	s.app.SetReconciler(adapter)

	log.Println("[fugo] flutter client connected")

	for {
		event, err := stream.Recv()
		if err != nil {
			log.Printf("[fugo] client disconnected: %v", err)

			return err
		}

		s.app.HandleEvent(event)
	}
}

// StartServer listens on addr, registers the FugoRender and gRPC health
// services with keepalive enabled, and serves in a background goroutine. It
// uses a Unix domain socket when addr has no colon (chmod 0600) and TCP
// otherwise, returning the running server and listener for later shutdown.
func StartServer(addr string, app AppHandler) (*grpc.Server, net.Listener, error) {
	network, address := resolveNetwork(addr)

	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, nil, fmt.Errorf("listen %s: %w", address, err)
	}

	if network == "unix" {
		if err := os.Chmod(address, 0o600); err != nil {
			return nil, nil, fmt.Errorf("chmod %s: %w", address, err)
		}
	}

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	server := newKeepaliveServer()
	fugov1.RegisterFugoRenderServer(server, NewServer(app))
	healthpb.RegisterHealthServer(server, healthSrv)

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("[fugo] server stopped: %v", err)
		}
	}()

	return server, listener, nil
}

func resolveNetwork(addr string) (string, string) {
	if isUDS(addr) {
		// Remove any stale socket file left by a previous run so net.Listen
		// can bind; a missing or already-removed file is fine to ignore.
		_ = os.Remove(addr)

		return "unix", addr
	}

	return "tcp", addr
}

func isUDS(addr string) bool {
	for _, c := range addr {
		if c == ':' {
			return false
		}
	}

	return true
}

func newKeepaliveServer() *grpc.Server {
	return grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    10 * time.Second,
			Timeout: 3 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)
}

type grpcStreamAdapter struct {
	stream fugov1.FugoRender_RenderStreamServer
}

func (a *grpcStreamAdapter) Send(payload *fugov1.RenderPayload) error {
	return a.stream.Send(payload)
}
