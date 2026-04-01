package handlers

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"auth/internal/infrastructure/validator"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VerifyEmailHandler struct {
	verifyEmailAction *actions.VerifyEmail
	env               config.Env
}

func NewVerifyEmail(userRepository userModel.Repository, jwtService *services.JWTService, env config.Env) *VerifyEmailHandler {
	return &VerifyEmailHandler{
		verifyEmailAction: actions.NewVerifyEmail(userRepository, jwtService),
		env:               env,
	}
}

func (handler *VerifyEmailHandler) Handle(context *gin.Context) {
	var verifyData data.VerifyEmail
	if validator.New(context, handler.env).Fails(&verifyData) {
		return
	}

	err := handler.verifyEmailAction.Execute(verifyData.Token)

	if err != nil {
		exceptions.NewUnprocessableEntity(context, handler.env).Throw(err)
		return
	}

	context.Status(http.StatusNoContent)
}
