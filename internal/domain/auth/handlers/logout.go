package handlers

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"auth/internal/infrastructure/redis"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogoutHandler struct {
	logoutAction *actions.Logout
	env          config.Env
}

func NewLogout(redisRepository *redis.Repository, env config.Env) *LogoutHandler {
	return &LogoutHandler{
		logoutAction: actions.NewLogout(redisRepository),
		env:          env,
	}
}

func (handler *LogoutHandler) Handle(context *gin.Context) {
	refreshToken, err := context.Cookie("refresh_token")

	if err != nil {
		context.Status(http.StatusNoContent)
		return
	}

	err = handler.logoutAction.Execute(context.Request.Context(), refreshToken)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	context.SetSameSite(http.SameSiteStrictMode)
	context.SetCookie("refresh_token", "", -1, "/", "", false, true)
	context.Status(http.StatusNoContent)
}
