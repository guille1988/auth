package responses

import (
	"api/internal/domain/auth/services"
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

func (response *LoginResponse) Make(data *services.TokenResponse) {
	response.Context.SetSameSite(http.SameSiteStrictMode)
	response.Context.SetCookie("refresh_token", data.RefreshToken, 3600*24*7, "/", "", false, true)

	response.Context.JSON(http.StatusOK, loginResponseData{
		AccessToken: data.AccessToken,
		TokenType:   data.TokenType,
		ExpiresIn:   data.ExpiresIn,
	})
}
