package grpc

import (
	authgrpc "auth/internal/domain/auth/grpc"
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/config"
	"context"
	"testing"
	"time"

	authv1 "github.com/guille1988/go-app-shared/rpc/auth/v1"

	"github.com/stretchr/testify/assert"
)

func testAuthConfig(accessTokenExpire time.Duration) config.AuthConfig {
	return config.AuthConfig{
		JWTSecret:               "secret",
		AccessTokenExpire:       accessTokenExpire,
		EmailVerificationExpire: time.Hour,
	}
}

func validate(server *authgrpc.Server, token string) *authv1.ValidateTokenResponse {
	response, _ := server.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: token})
	return response
}

func TestValidateToken(test *testing.T) {
	jwtService := services.NewJWTService(testAuthConfig(time.Hour))
	server := authgrpc.NewServer(jwtService)

	test.Run("it should accept a valid access token and return its user uuid", func(test *testing.T) {
		userUUID := "user-uuid-123"
		tokens, err := jwtService.GenerateAccessToken(userUUID, "refresh-token")
		assert.NoError(test, err)

		response, err := server.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: tokens.AccessToken})
		assert.NoError(test, err)
		assert.True(test, response.GetValid())
		assert.Equal(test, userUUID, response.GetUserUuid())
	})

	test.Run("it should reject an expired token with the EXPIRED reason", func(test *testing.T) {
		expiredService := services.NewJWTService(testAuthConfig(-time.Minute))
		tokens, err := expiredService.GenerateAccessToken("user-uuid-123", "refresh-token")
		assert.NoError(test, err)

		response := validate(server, tokens.AccessToken)
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_EXPIRED, response.GetReason())
	})

	test.Run("it should reject a token issued for another purpose with the WRONG_PURPOSE reason", func(test *testing.T) {
		verificationToken, err := jwtService.GenerateEmailVerificationToken("user-uuid-123")
		assert.NoError(test, err)

		response := validate(server, verificationToken)
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_WRONG_PURPOSE, response.GetReason())
	})

	test.Run("it should reject a token signed with another secret with the MALFORMED reason", func(test *testing.T) {
		otherService := services.NewJWTService(config.AuthConfig{JWTSecret: "another-secret", AccessTokenExpire: time.Hour})
		tokens, err := otherService.GenerateAccessToken("user-uuid-123", "refresh-token")
		assert.NoError(test, err)

		response := validate(server, tokens.AccessToken)
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_MALFORMED, response.GetReason())
	})

	test.Run("it should reject garbage with the MALFORMED reason", func(test *testing.T) {
		response := validate(server, "not-a-jwt")
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_MALFORMED, response.GetReason())
	})

	test.Run("it should reject an empty token with the MALFORMED reason", func(test *testing.T) {
		response := validate(server, "")
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_MALFORMED, response.GetReason())
	})

	test.Run("it should never surface an invalid token as a transport error", func(test *testing.T) {
		for _, token := range []string{"", "not-a-jwt"} {
			_, err := server.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: token})
			assert.NoError(test, err)
		}
	})
}
