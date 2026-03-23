package responses

import (
	"api/internal/domain/auth/services"
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

func (response *RefreshResponse) Make(data *services.TokenResponse) {
	response.Context.SetSameSite(http.SameSiteStrictMode)
	response.Context.SetCookie("refresh_token", data.RefreshToken, 3600*24*7, "/", "", false, true)

	response.Context.JSON(http.StatusOK, refreshResponseData{
		AccessToken: data.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   data.ExpiresIn,
	})
}
