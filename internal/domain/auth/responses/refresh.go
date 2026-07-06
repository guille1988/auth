package responses

import (
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RefreshResponse struct {
	Context *gin.Context
}

func NewRefreshResponse(context *gin.Context) *RefreshResponse {
	return &RefreshResponse{Context: context}
}

type refreshResponseData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (response *RefreshResponse) Make(data *services.TokenResponse, authConfig config.AuthConfig, env config.Env) {
	response.Context.SetSameSite(http.SameSiteStrictMode)
	maxAge := int(authConfig.RefreshTokenExpire.Seconds())
	secure := env == config.ProductionEnv
	response.Context.SetCookie("refresh_token", data.RefreshToken, maxAge, "/", "", secure, true)

	response.Context.JSON(http.StatusOK, refreshResponseData{
		AccessToken: data.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   data.ExpiresIn,
	})
}
