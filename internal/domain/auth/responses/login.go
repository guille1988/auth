package responses

import (
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginResponse struct {
	Context *gin.Context
}

func NewLoginResponse(context *gin.Context) *LoginResponse {
	return &LoginResponse{Context: context}
}

type loginResponseData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (response *LoginResponse) Make(data *services.TokenResponse, authConfig config.AuthConfig, env config.Env) {
	response.Context.SetSameSite(http.SameSiteStrictMode)
	maxAge := int(authConfig.RefreshTokenExpire.Seconds())
	secure := env == config.ProductionEnv
	response.Context.SetCookie("refresh_token", data.RefreshToken, maxAge, "/", "", secure, true)

	response.Context.JSON(http.StatusOK, loginResponseData{
		AccessToken: data.AccessToken,
		TokenType:   data.TokenType,
		ExpiresIn:   data.ExpiresIn,
	})
}
