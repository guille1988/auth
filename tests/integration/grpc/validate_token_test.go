package grpc

import (
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	grpcprovider "auth/internal/infrastructure/providers/grpc"
	"auth/tests/integration"
	"context"
	"net"
	"testing"
	"time"

	authv1 "github.com/guille1988/go-app-shared/rpc/auth/v1"

	"github.com/stretchr/testify/assert"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

/*
newValidationClient serves the real, fully wired gRPC server on an
in-memory bufconn listener and returns a client dialed against it.
*/
func newValidationClient(test *testing.T) authv1.AuthServiceClient {
	appInstance, err := integration.GetApp()
	assert.NoError(test, err)

	server := grpcprovider.NewServer(appInstance)
	listener := bufconn.Listen(1024 * 1024)

	go func() {
		_ = server.Serve(listener)
	}()

	conn, err := googlegrpc.NewClient(
		"passthrough:///bufnet",
		googlegrpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NoError(test, err)

	test.Cleanup(func() {
		_ = conn.Close()
		server.Stop()
	})

	return authv1.NewAuthServiceClient(conn)
}

func TestValidateTokenRPC(test *testing.T) {
	integration.TestCase(test, "it should validate a real user's access token and return their uuid", func(test *testing.T) {
		token := integration.GetToken()
		client := newValidationClient(test)

		response, err := client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: token})
		assert.NoError(test, err)
		assert.True(test, response.GetValid())

		appInstance, err := integration.GetApp()
		assert.NoError(test, err)

		userRepo := userModel.NewRepository(appInstance.Container.DefaultConnection)
		user, err := userRepo.FindByEmail("test@example.com")
		assert.NoError(test, err)
		assert.Equal(test, user.UUID.String(), response.GetUserUuid())
	})

	integration.TestCase(test, "it should report an expired token as invalid with the EXPIRED reason", func(test *testing.T) {
		client := newValidationClient(test)

		expiredConfig := config.AuthConfig{
			JWTSecret:         integration.TestConfig.Auth.JWTSecret,
			AccessTokenExpire: -time.Minute,
		}
		tokens, err := services.NewJWTService(expiredConfig).GenerateAccessToken("user-uuid-123", "refresh-token")
		assert.NoError(test, err)

		response, err := client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: tokens.AccessToken})
		assert.NoError(test, err)
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_EXPIRED, response.GetReason())
	})
}
