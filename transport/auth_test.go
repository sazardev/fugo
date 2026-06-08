package transport

import (
	"context"
	"testing"
	"time"

	"github.com/sazardev/fugo/engine"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type authStubApp struct{}

func (authStubApp) HandleEvent(*fugov1.ClientEvent)   {}
func (authStubApp) SetReconciler(engine.RenderStream) {}

// TestAuthTokenInterceptor verifies the opt-in per-run token: with FUGO_TOKEN
// set, the render stream rejects clients that present no token or a wrong one,
// and admits the matching token.
func TestAuthTokenInterceptor(t *testing.T) {
	t.Setenv("FUGO_TOKEN", "secret-123")

	srv, lis, err := StartServer("127.0.0.1:0", authStubApp{})
	if err != nil {
		t.Fatalf("StartServer: %v", err)
	}
	defer srv.GracefulStop()

	addr := lis.Addr().String()

	openStream := func(token string) error {
		conn, dialErr := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if dialErr != nil {
			return dialErr
		}
		defer func() { _ = conn.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if token != "" {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(tokenMetadataKey, token))
		}

		stream, streamErr := fugov1.NewFugoRenderClient(conn).RenderStream(ctx)
		if streamErr != nil {
			return streamErr
		}

		_, recvErr := stream.Recv() // first read triggers the server interceptor

		return recvErr
	}

	if got := status.Code(openStream("")); got != codes.Unauthenticated {
		t.Errorf("missing token: code = %v, want Unauthenticated", got)
	}

	if got := status.Code(openStream("wrong")); got != codes.Unauthenticated {
		t.Errorf("wrong token: code = %v, want Unauthenticated", got)
	}

	// A correct token is admitted; the stream then blocks until the deadline,
	// so any code other than Unauthenticated means auth passed.
	if got := status.Code(openStream("secret-123")); got == codes.Unauthenticated {
		t.Errorf("correct token was rejected: code = %v", got)
	}
}
