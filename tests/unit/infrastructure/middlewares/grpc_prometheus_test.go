package middlewares

import (
	"auth/internal/infrastructure/middlewares"
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGRPCPrometheus_PassesThroughResponseAndCountsOK(t *testing.T) {
	interceptor := middlewares.GRPCPrometheus()
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/TestOK"}
	handler := func(_ context.Context, req any) (any, error) {
		return "response", nil
	}

	resp, err := interceptor(context.Background(), "request", info, handler)

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
	assert.Equal(t, float64(1), testutil.ToFloat64(
		middlewares.GRPCRequestsTotal.WithLabelValues(info.FullMethod, codes.OK.String()),
	))
}

func TestGRPCPrometheus_PassesThroughErrorAndCountsCode(t *testing.T) {
	interceptor := middlewares.GRPCPrometheus()
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/TestUnauthenticated"}
	wantErr := status.Error(codes.Unauthenticated, "invalid token")
	handler := func(_ context.Context, req any) (any, error) {
		return nil, wantErr
	}

	resp, err := interceptor(context.Background(), "request", info, handler)

	assert.Nil(t, resp)
	assert.Equal(t, wantErr, err)
	assert.Equal(t, float64(1), testutil.ToFloat64(
		middlewares.GRPCRequestsTotal.WithLabelValues(info.FullMethod, codes.Unauthenticated.String()),
	))
}
