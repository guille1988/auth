package exceptions

import (
	"auth/internal/infrastructure/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewServer(context *gin.Context, env config.Env) *Exception {
	return newException(context, env, http.StatusInternalServerError, false)
}
