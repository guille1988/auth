package middlewares

import (
	"auth/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterMiddlewares(engine *gin.Engine, env config.Env) {
	engine.Use(gin.Recovery())
	engine.Use(Logger())
	engine.Use(IgnoreFavicon(env))
	engine.Use(Prometheus())
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
