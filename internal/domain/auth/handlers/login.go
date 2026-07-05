package handlers

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/responses"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"auth/internal/infrastructure/redis"
	"auth/internal/infrastructure/validator"
	"errors"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	loginAction *actions.Login
	env         config.Env
}

func NewLogin(redisRepository *redis.Repository, publisher actions.MessagePublisher, userRepository userModel.Repository, jwtService *services.JWTService, authConfig config.AuthConfig, env config.Env) *LoginHandler {
	return &LoginHandler{
		loginAction: actions.NewLogin(userRepository, redisRepository, jwtService, authConfig, publisher),
		env:         env,
	}
}

func (handler *LoginHandler) Handle(context *gin.Context) {
	var loginData data.Login
	if validator.New(context, handler.env).Fails(&loginData) {
		return
	}

	response, err := handler.loginAction.Execute(context.Request.Context(), loginData, context.GetHeader("User-Agent"))

	if err != nil {
		if errors.Is(err, actions.ErrEmailNotVerified) {
			exceptions.NewForbidden(context, handler.env).Throw(err)
			return
		}

		exceptions.NewUnauthorized(context, handler.env).Throw(err)
		return
	}

	responses.NewLoginResponse(context).Make(response)
}
