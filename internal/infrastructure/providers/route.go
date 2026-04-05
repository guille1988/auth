package providers

import (
	"auth/internal/domain/auth"
	"auth/internal/domain/health"
	"auth/internal/domain/user"
	"auth/internal/infrastructure/app"

	"github.com/gin-gonic/gin"
)

// RouteRegister is the interface for registering routes in a module.
type RouteRegister interface {
	Register(group *gin.RouterGroup)
}

// RegisterRoutes handles the wiring of dependencies and route registration.
func RegisterRoutes(engine *gin.Engine, app *app.App) {
	api := engine.Group("/api")

	registers := []RouteRegister{
		health.NewModule(),
		auth.NewModule(app.Container.DefaultConnection, app.Container.Redis, app.Container.RabbitMQProvider, app.Config.Auth, app.Config.App.Env),
		user.NewModule(app.Container.DefaultConnection, *app.Config),
	}

	for _, register := range registers {
		register.Register(api)
	}
}
