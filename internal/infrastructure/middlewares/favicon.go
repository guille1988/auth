package middlewares

import (
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"

	"github.com/gin-gonic/gin"
)

func IgnoreFavicon(env config.Env) gin.HandlerFunc {
	return func(context *gin.Context) {
		if context.Request.RequestURI == "/favicon.ico" {
			exceptions.NewNotFound(context, env).Throw(nil)

			return
		}
		context.Next()
	}
}
