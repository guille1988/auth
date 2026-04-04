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

type ResendVerificationEmailHandler struct {
	resendAction *actions.ResendVerificationEmail
	env          config.Env
}

func NewResendVerificationEmail(userRepository userModel.Repository, jwtService *services.JWTService, publisher actions.MessagePublisher, authConfig config.AuthConfig, env config.Env) *ResendVerificationEmailHandler {
	return &ResendVerificationEmailHandler{
		resendAction: actions.NewResendVerificationEmail(userRepository, jwtService, publisher, authConfig),
		env:          env,
	}
}

func (handler *ResendVerificationEmailHandler) Handle(context *gin.Context) {
	var resendData data.ResendVerification
	if validator.New(context, handler.env).Fails(&resendData) {
		return
	}

	err := handler.resendAction.Execute(context.Request.Context(), resendData.Email)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	context.Status(http.StatusNoContent)
}
