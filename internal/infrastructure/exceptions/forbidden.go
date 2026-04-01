package exceptions

import (
	"auth/internal/infrastructure/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewForbidden(context *gin.Context, env config.Env) *Exception {
	return newException(context, env, http.StatusForbidden, false)
}
