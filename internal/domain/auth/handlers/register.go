package handlers

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/responses"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"auth/internal/infrastructure/providers/messaging"
	"auth/internal/infrastructure/redis"
	"auth/internal/infrastructure/validator"
	"errors"

	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	userRepository userModel.Repository
	registerAction *actions.Register
	env            config.Env
}

func NewRegister(redisRepository *redis.Repository, rabbitMQProvider *messaging.RabbitMQRegister, userRepository userModel.Repository, jwtService *services.JWTService, authConfig config.AuthConfig, env config.Env) *RegisterHandler {
	return &RegisterHandler{
		userRepository: userRepository,
		registerAction: actions.NewRegister(userRepository, redisRepository, rabbitMQProvider, jwtService, authConfig),
		env:            env,
	}
}

func (handler *RegisterHandler) Handle(context *gin.Context) {
	var registerData data.Register
	if validator.New(context, handler.env).Fails(&registerData) {
		return
	}

	exists, err := handler.userRepository.ExistByEmail(registerData.Email)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)

		return
	}

	if exists {
		exceptions.NewUnprocessableEntity(context, handler.env).
			Throw(errors.New("this e-mail already exists"))

		return
	}

	var response *services.TokenResponse
	response, err = handler.registerAction.Execute(registerData, context.GetHeader("User-Agent"))

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)

		return
	}

	responses.NewRegisterResponse(context).Make(response)
}
