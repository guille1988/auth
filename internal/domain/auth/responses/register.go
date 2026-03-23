package responses

import (
	"api/internal/domain/auth/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterResponse struct {
	Context *gin.Context
}

func NewRegisterResponse(context *gin.Context) *RegisterResponse {
	return &RegisterResponse{Context: context}
}

type registerResponseData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (response *RegisterResponse) Make(data *services.TokenResponse) {
	response.Context.SetSameSite(http.SameSiteStrictMode)
	response.Context.SetCookie("refresh_token", data.RefreshToken, 3600*24*7, "/", "", false, true)

	response.Context.JSON(http.StatusCreated, registerResponseData{
		AccessToken: data.AccessToken,
		TokenType:   data.TokenType,
		ExpiresIn:   data.ExpiresIn,
	})
}
