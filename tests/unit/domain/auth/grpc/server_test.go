package grpc

import (
	authgrpc "auth/internal/domain/auth/grpc"
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/config"
	"context"
	"errors"
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

type fakeSessionChecker struct {
	live  bool
	err   error
	calls int
}

func (fake *fakeSessionChecker) HasLiveSession(_ context.Context, _ string) (bool, error) {
	fake.calls++
	return fake.live, fake.err
}

func validate(server *authgrpc.Server, token string) *authv1.ValidateTokenResponse {
	response, _ := server.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: token})
	return response
}

func TestValidateToken(test *testing.T) {
	jwtService := services.NewJWTService(testAuthConfig(time.Hour))
	server := authgrpc.NewServer(jwtService, &fakeSessionChecker{live: true})

	test.Run("it should accept a valid access token and return its user uuid", func(test *testing.T) {
		userUUID := "user-uuid-123"
		tokens, err := jwtService.GenerateAccessToken(userUUID, "refresh-token")
		assert.NoError(test, err)

		var response *authv1.ValidateTokenResponse
		response, err = server.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: tokens.AccessToken})
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

	test.Run("it should reject a JWT-valid token whose user has no live session with the REVOKED reason", func(test *testing.T) {
		revokedServer := authgrpc.NewServer(jwtService, &fakeSessionChecker{live: false})
		tokens, err := jwtService.GenerateAccessToken("user-uuid-123", "refresh-token")
		assert.NoError(test, err)

		var response *authv1.ValidateTokenResponse
		response, err = revokedServer.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: tokens.AccessToken})

		assert.NoError(test, err)
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_REVOKED, response.GetReason())
	})

	test.Run("it should surface a session-store failure as a transport error, not a token verdict", func(test *testing.T) {
		brokenServer := authgrpc.NewServer(jwtService, &fakeSessionChecker{err: errors.New("redis down")})
		tokens, err := jwtService.GenerateAccessToken("user-uuid-123", "refresh-token")
		assert.NoError(test, err)

		var response *authv1.ValidateTokenResponse
		response, err = brokenServer.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: tokens.AccessToken})

		assert.Error(test, err)
		assert.Nil(test, response)
	})

	test.Run("it should not consult the session store for tokens that fail JWT validation", func(test *testing.T) {
		recorder := &fakeSessionChecker{live: true}
		recordingServer := authgrpc.NewServer(jwtService, recorder)

		expiredTokens, err := services.NewJWTService(testAuthConfig(-time.Minute)).GenerateAccessToken("user-uuid-123", "refresh-token")
		assert.NoError(test, err)

		var verificationToken string
		verificationToken, err = jwtService.GenerateEmailVerificationToken("user-uuid-123")
		assert.NoError(test, err)

		for _, token := range []string{"", "not-a-jwt", expiredTokens.AccessToken, verificationToken} {
			_, _ = recordingServer.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: token})
		}

		assert.Equal(test, 0, recorder.calls)
	})
}
