package exceptions

import (
	"api/internal/infrastructure/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewNotFound(context *gin.Context, env config.Env) *Exception {
	return newException(context, env, http.StatusNotFound, false)
}
