package middlewares

import (
	"api/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

func RegisterMiddlewares(engine *gin.Engine, env config.Env) {
	engine.Use(gin.Recovery())
	engine.Use(Logger())
	engine.Use(IgnoreFavicon(env))
}
