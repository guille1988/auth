package handlers

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/domain/auth/responses"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"auth/internal/infrastructure/redis"

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
