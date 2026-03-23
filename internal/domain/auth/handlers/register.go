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
	"errors"

	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	userRepository userModel.Repository
	registerAction *actions.Register
	env            config.Env
}

func NewRegister(redisRepository *redis.Repository, userRepository userModel.Repository, jwtService *services.JWTService, authConfig config.AuthConfig, env config.Env) *RegisterHandler {
	return &RegisterHandler{
		userRepository: userRepository,
		registerAction: actions.NewRegister(userRepository, redisRepository, jwtService, authConfig),
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
