package stress

import (
	"auth/internal/domain/stress/actions"
	"auth/internal/domain/stress/handlers"
	"auth/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

type Module struct {
	publisher actions.MessagePublisher
	env       config.Env
}

func NewModule(publisher actions.MessagePublisher, env config.Env) *Module {
	return &Module{
		publisher: publisher,
		env:       env,
	}
}

func (module *Module) Register(group *gin.RouterGroup) {
	stress := group.Group("/stress")
	{
		stress.POST("", handlers.NewStress(module.publisher, module.env).Handle)
	}
}
