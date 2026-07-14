package middlewares

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var GRPCRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests",
	},
	[]string{"method", "code"},
)

func GRPCPrometheus() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)

		GRPCRequestsTotal.WithLabelValues(info.FullMethod, status.Code(err).String()).Inc()

		return resp, err
	}
}
