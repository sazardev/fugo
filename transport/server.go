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

type Server struct {
	fugov1.UnimplementedFugoRenderServer
	app AppHandler
}

type AppHandler interface {
	HandleEvent(ev *fugov1.ClientEvent)
	SetReconciler(stream engine.RenderStream)
}

func NewServer(app AppHandler) *Server {
	return &Server{app: app}
}

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
		os.Remove(addr)

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
