package handlers

import (
	"auth/internal/domain/stress/actions"
	"auth/internal/domain/stress/data"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"auth/internal/infrastructure/validator"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StressHandler struct {
	action *actions.SendStress
	env    config.Env
}

func NewStress(publisher actions.MessagePublisher, env config.Env) *StressHandler {
	return &StressHandler{
		action: actions.NewSendStress(publisher),
		env:    env,
	}
}

func (handler *StressHandler) Handle(ctx *gin.Context) {
	var stressData data.Stress
	if validator.New(ctx, handler.env).Fails(&stressData) {
		return
	}

	if err := handler.action.Execute(ctx.Request.Context(), stressData); err != nil {
		exceptions.NewServer(ctx, handler.env).Throw(err)
		return
	}

	ctx.Status(http.StatusAccepted)
}
