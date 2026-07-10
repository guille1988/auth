package responses

import (
	"auth/internal/domain/auth/responses"
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/config"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest("POST", "/api/auth/login", nil)
	return ginContext, recorder
}

func refreshCookieFrom(recorder *httptest.ResponseRecorder) (cookie *http.Cookie, found bool) {
	for _, c := range recorder.Result().Cookies() {
		if c.Name == "refresh_token" {
			return c, true
		}
	}
	return nil, false
}

func TestLoginResponseCookieHygiene(test *testing.T) {
	test.Run("it should derive the cookie max-age from AuthConfig, not a hardcoded value", func(test *testing.T) {
		ginContext, recorder := newTestGinContext()

		authConfig := config.AuthConfig{RefreshTokenExpire: 1 * time.Hour}
		responses.NewLoginResponse(ginContext).Make(&services.TokenResponse{
			AccessToken:  "access",
			RefreshToken: "refresh",
			TokenType:    "Bearer",
			ExpiresIn:    900,
		}, authConfig, config.LocalEnv)

		cookie, found := refreshCookieFrom(recorder)
		assert.True(test, found, "refresh_token cookie must be set")
		assert.Equal(test, 3600, cookie.MaxAge, "max-age must match AuthConfig.RefreshTokenExpire (1h), not a hardcoded 7 days")
	})

	test.Run("it should mark the cookie Secure in production", func(test *testing.T) {
		ginContext, recorder := newTestGinContext()

		authConfig := config.AuthConfig{RefreshTokenExpire: 1 * time.Hour}
		responses.NewLoginResponse(ginContext).Make(&services.TokenResponse{RefreshToken: "refresh"}, authConfig, config.ProductionEnv)

		cookie, found := refreshCookieFrom(recorder)
		assert.True(test, found)
		assert.True(test, cookie.Secure, "refresh_token cookie must be Secure in production")
	})

	test.Run("it should not mark the cookie Secure in local", func(test *testing.T) {
		ginContext, recorder := newTestGinContext()

		authConfig := config.AuthConfig{RefreshTokenExpire: 1 * time.Hour}
		responses.NewLoginResponse(ginContext).Make(&services.TokenResponse{RefreshToken: "refresh"}, authConfig, config.LocalEnv)

		cookie, found := refreshCookieFrom(recorder)
		assert.True(test, found)
		assert.False(test, cookie.Secure, "refresh_token cookie must not be Secure over plain HTTP in local dev")
	})
}
