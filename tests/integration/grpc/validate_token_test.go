package grpc

import (
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/app"
	"auth/internal/infrastructure/config"
	grpcprovider "auth/internal/infrastructure/providers/grpc"
	"auth/tests/integration"
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	authv1 "github.com/guille1988/go-app-shared/rpc/auth/v1"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

/*
the newValidationClient serves the real, fully wired gRPC server on an
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

	var conn *googlegrpc.ClientConn
	conn, err = googlegrpc.NewClient(
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

// session bundles the credentials one register/login produces.
type session struct {
	accessToken   string
	refreshCookie *http.Cookie
}

func executeJSON(_ *testing.T, method, path string, payload map[string]string) (*http.Response, map[string]any) {
	body, _ := json.Marshal(payload)
	request, _ := http.NewRequest(method, path, bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")

	recorder := integration.ExecuteRequest(request)

	var data map[string]any
	_ = json.Unmarshal(recorder.Body.Bytes(), &data)

	return recorder.Result(), data
}

func refreshCookieFrom(test *testing.T, response *http.Response) *http.Cookie {
	for _, cookie := range response.Cookies() {
		if cookie.Name == "refresh_token" {
			return cookie
		}
	}

	test.Fatal("refresh_token cookie not found")
	return nil
}

/*
registerVerifiedUser registers a fresh user, verifies their email, and
returns their credentials plus the first session (flow from auth_test.go).
*/
func registerVerifiedUser(test *testing.T) (email, password string, firstSession session) {
	email = gofakeit.Email()
	password = "password123"

	response, data := executeJSON(test, "POST", "/api/auth/register", map[string]string{
		"name": gofakeit.Name(), "email": email, "password": password,
	})
	assert.Equal(test, http.StatusCreated, response.StatusCode)

	appInstance, err := integration.GetApp()
	assert.NoError(test, err)

	userRepository := userModel.NewRepository(appInstance.Container.DefaultConnection)
	var user *userModel.User
	user, err = userRepository.FindByEmail(email)
	assert.NoError(test, err)

	var verificationToken string
	verificationToken, err = services.NewJWTService(appInstance.Config.Auth).GenerateEmailVerificationToken(user.UUID.String())
	assert.NoError(test, err)

	verifyResponse, _ := executeJSON(test, "POST", "/api/auth/verify-email", map[string]string{"token": verificationToken})
	assert.Equal(test, http.StatusNoContent, verifyResponse.StatusCode)

	return email, password, session{
		accessToken:   data["access_token"].(string),
		refreshCookie: refreshCookieFrom(test, response),
	}
}

func login(test *testing.T, email, password string) session {
	response, data := executeJSON(test, "POST", "/api/auth/login", map[string]string{"email": email, "password": password})
	assert.Equal(test, http.StatusOK, response.StatusCode)

	return session{
		accessToken:   data["access_token"].(string),
		refreshCookie: refreshCookieFrom(test, response),
	}
}

func logout(test *testing.T, refreshCookie *http.Cookie) {
	request, _ := http.NewRequest("DELETE", "/api/auth/logout", nil)
	request.AddCookie(refreshCookie)

	recorder := integration.ExecuteRequest(request)
	assert.Equal(test, http.StatusNoContent, recorder.Code)
}

func validateOverWire(test *testing.T, client authv1.AuthServiceClient, accessToken string) *authv1.ValidateTokenResponse {
	response, err := client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: accessToken})
	assert.NoError(test, err)

	return response
}

func TestValidateTokenRPC(test *testing.T) {
	integration.TestCase(test, "it should validate a real user's access token and return their uuid", func(test *testing.T) {
		token := integration.GetToken()
		client := newValidationClient(test)

		response, err := client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: token})
		assert.NoError(test, err)
		assert.True(test, response.GetValid())

		var appInstance *app.App
		appInstance, err = integration.GetApp()
		assert.NoError(test, err)

		userRepo := userModel.NewRepository(appInstance.Container.DefaultConnection)
		var user *userModel.User
		user, err = userRepo.FindByEmail("test@example.com")
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

		var response *authv1.ValidateTokenResponse
		response, err = client.ValidateToken(context.Background(), &authv1.ValidateTokenRequest{Token: tokens.AccessToken})
		assert.NoError(test, err)
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_EXPIRED, response.GetReason())
	})
}

func TestSessionAwareValidation(test *testing.T) {
	integration.TestCase(test, "it should report REVOKED after the user's only session is logged out", func(test *testing.T) {
		_, _, userSession := registerVerifiedUser(test)
		client := newValidationClient(test)

		assert.True(test, validateOverWire(test, client, userSession.accessToken).GetValid())

		logout(test, userSession.refreshCookie)

		response := validateOverWire(test, client, userSession.accessToken)
		assert.False(test, response.GetValid())
		assert.Equal(test, authv1.ValidateTokenResponse_REVOKED, response.GetReason())
	})

	integration.TestCase(test, "it should keep tokens valid while another session of the same user lives", func(test *testing.T) {
		email, password, firstSession := registerVerifiedUser(test)
		secondSession := login(test, email, password)
		client := newValidationClient(test)

		logout(test, firstSession.refreshCookie)

		// Per-user semantics: the second device keeps every token alive.
		assert.True(test, validateOverWire(test, client, firstSession.accessToken).GetValid())
		assert.True(test, validateOverWire(test, client, secondSession.accessToken).GetValid())

		logout(test, secondSession.refreshCookie)

		assert.Equal(test, authv1.ValidateTokenResponse_REVOKED, validateOverWire(test, client, firstSession.accessToken).GetReason())
		assert.Equal(test, authv1.ValidateTokenResponse_REVOKED, validateOverWire(test, client, secondSession.accessToken).GetReason())
	})

	integration.TestCase(test, "it should keep the user's token valid across a refresh rotation", func(test *testing.T) {
		_, _, userSession := registerVerifiedUser(test)
		client := newValidationClient(test)

		refreshRequest, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
		refreshRequest.AddCookie(userSession.refreshCookie)
		recorder := integration.ExecuteRequest(refreshRequest)
		assert.Equal(test, http.StatusOK, recorder.Code)

		var refreshed map[string]any
		_ = json.Unmarshal(recorder.Body.Bytes(), &refreshed)

		assert.True(test, validateOverWire(test, client, userSession.accessToken).GetValid(), "the pre-rotation access token stays valid while its user has a live session")
		assert.True(test, validateOverWire(test, client, refreshed["access_token"].(string)).GetValid())
	})
}
