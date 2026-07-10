package bootstrap

import (
	"auth/internal/infrastructure/app"
	grpcprovider "auth/internal/infrastructure/providers/grpc"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
)

/*
startGRPCServer builds the gRPC server, binds its listener, and serves in a
goroutine that reports real failures to serverErrors. The caller owns the
lifecycle: it must stop the returned server through shutdownGRPCServer (or
Stop on the error path).
*/
func startGRPCServer(appInstance *app.App, serverErrors chan error) (*grpc.Server, error) {
	grpcServer := grpcprovider.NewServer(appInstance)
	grpcAddr := fmt.Sprintf("%s:%s", appInstance.Config.App.Host, appInstance.Config.App.GRPCPort)

	listener, err := net.Listen("tcp", grpcAddr)

	if err != nil {
		return nil, fmt.Errorf("grpc listen error: %w", err)
	}

	go func() {
		slog.Info("grpc server is starting", "addr", grpcAddr)

		// Serve returns nil after GracefulStop/Stop, so only real failures land here.
		if serveErr := grpcServer.Serve(listener); serveErr != nil {
			serverErrors <- serveErr
		}
	}()

	return grpcServer, nil
}

/*
shutdownGRPCServer concern: Specific logic to stop the gRPC server gracefully.
GracefulStop blocks until every in-flight RPC finishes, which can be forever
on a stuck stream, so it falls back to a hard Stop after 30s.
*/
func shutdownGRPCServer(grpcServer *grpc.Server) {
	done := make(chan struct{})

	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		slog.Warn("grpc graceful stop timed out, forcing stop")
		grpcServer.Stop()
	}
}
