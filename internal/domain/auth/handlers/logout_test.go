package handlers

import (
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/redis"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestLogoutContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)

	request := httptest.NewRequest("DELETE", "/api/auth/logout", nil)
	request.AddCookie(&http.Cookie{Name: "refresh_token", Value: "some-refresh-token"})
	ginContext.Request = request

	return ginContext, recorder
}

func cookieNamed(recorder *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func TestLogoutClearsCookieSecureByEnv(test *testing.T) {
	client := goredis.NewClient(&goredis.Options{Addr: "redis:6379", Password: "auth"})
	defer func() { _ = client.Close() }()
	repository := redis.NewRepository(client)

	test.Run("it should mark the cleared cookie Secure in production", func(test *testing.T) {
		ctx, recorder := newTestLogoutContext()
		NewLogout(repository, config.ProductionEnv).Handle(ctx)

		cookie := cookieNamed(recorder, "refresh_token")
		if assert.NotNil(test, cookie) {
			assert.True(test, cookie.Secure, "the cleared refresh_token cookie must be Secure in production")
		}
	})

	test.Run("it should not mark the cleared cookie Secure in local", func(test *testing.T) {
		ctx, recorder := newTestLogoutContext()
		NewLogout(repository, config.LocalEnv).Handle(ctx)

		cookie := cookieNamed(recorder, "refresh_token")

		if assert.NotNil(test, cookie) {
			assert.False(test, cookie.Secure, "the cleared refresh_token cookie must not be Secure over plain HTTP in local dev")
		}
	})
}
