package handlers

import (
	"api/internal/domain/auth/actions"
	"api/internal/domain/auth/data"
	"api/internal/domain/auth/responses"
	"api/internal/domain/auth/services"
	userModel "api/internal/domain/user/model"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/exceptions"
	"api/internal/infrastructure/redis"
	"api/internal/infrastructure/validator"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	loginAction *actions.Login
	env         config.Env
}

func NewLogin(redisRepository *redis.Repository, userRepository userModel.Repository, jwtService *services.JWTService, authConfig config.AuthConfig, env config.Env) *LoginHandler {
	return &LoginHandler{
		loginAction: actions.NewLogin(userRepository, redisRepository, jwtService, authConfig),
		env:         env,
	}
}

func (handler *LoginHandler) Handle(context *gin.Context) {
	var loginData data.Login
	if validator.New(context, handler.env).Fails(&loginData) {
		return
	}

	response, err := handler.loginAction.Execute(loginData, context.GetHeader("User-Agent"))

	if err != nil {
		exceptions.NewUnauthorized(context, handler.env).Throw(err)
		return
	}

	responses.NewLoginResponse(context).Make(response)
}
