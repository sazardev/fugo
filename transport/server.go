package transport

import (
	"log"
	"net"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"

	"github.com/sazardev/fugo/engine"
	"google.golang.org/grpc"
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
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	server := grpc.NewServer()
	fugov1.RegisterFugoRenderServer(server, NewServer(app))

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Printf("[fugo] server stopped: %v", err)
		}
	}()

	return server, listener, nil
}

type grpcStreamAdapter struct {
	stream fugov1.FugoRender_RenderStreamServer
}

func (a *grpcStreamAdapter) Send(payload *fugov1.RenderPayload) error {
	return a.stream.Send(payload)
}
