package handlers

import (
	"api/internal/domain/auth/actions"
	"api/internal/domain/auth/responses"
	"api/internal/domain/auth/services"
	userModel "api/internal/domain/user/model"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/exceptions"
	"api/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
)

type RefreshHandler struct {
	refreshAction *actions.Refresh
	env           config.Env
}

func NewRefresh(redisRepository *redis.Repository, userRepository userModel.Repository, jwtService *services.JWTService, authConfig config.AuthConfig, env config.Env) *RefreshHandler {
	return &RefreshHandler{
		refreshAction: actions.NewRefresh(userRepository, redisRepository, jwtService, authConfig),
		env:           env,
	}
}

func (handler *RefreshHandler) Handle(context *gin.Context) {
	refreshToken, err := context.Cookie("refresh_token")

	if err != nil {
		exceptions.NewUnauthorized(context, handler.env).Throw(err)
		return
	}

	var response *services.TokenResponse
	response, err = handler.refreshAction.Execute(refreshToken)

	if err != nil {
		exceptions.NewUnauthorized(context, handler.env).Throw(err)
		return
	}

	responses.NewRefreshResponse(context).Make(response)
}
