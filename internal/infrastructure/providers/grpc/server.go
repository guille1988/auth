package grpc

import (
	authgrpc "auth/internal/domain/auth/grpc"
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/middlewares"
	"auth/internal/infrastructure/redis"

	authv1 "github.com/guille1988/go-app-shared/rpc/auth/v1"

	googlegrpc "google.golang.org/grpc"
)

/*
NewServer wires the domain gRPC handlers into a grpc.Server, the gRPC
counterpart of providers.RegisterRoutes.
*/
func NewServer(appInstance *app.App) *googlegrpc.Server {
	server := googlegrpc.NewServer(googlegrpc.UnaryInterceptor(middlewares.GRPCPrometheus()))

	sessionIndex := services.NewSessionIndex(redis.NewRepository(appInstance.Container.Redis))

	authv1.RegisterAuthServiceServer(server, authgrpc.NewServer(services.NewJWTService(appInstance.Config.Auth), sessionIndex))

	return server
}
