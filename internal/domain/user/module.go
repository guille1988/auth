package user

import (
	"auth/internal/domain/user/handlers"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	db     *gorm.DB
	env    config.Env
	config config.Config
}

func NewModule(db *gorm.DB, cfg config.Config) *Module {
	return &Module{
		db:     db,
		env:    cfg.App.Env,
		config: cfg,
	}
}

func (module *Module) Register(v1 *gin.RouterGroup) {
	group := v1.Group("/users")
	group.Use(middlewares.AuthMiddleware(module.config.Auth, module.env))
	group.Use(middlewares.EnsureEmailVerified(userModel.NewRepository(module.db), module.env))
	{
		group.GET("/", handlers.NewIndex(module.db, module.env).Handle)
		group.POST("/", handlers.NewStore(module.db, module.env).Handle)
		group.GET("/:uuid", handlers.NewShow(module.db, module.env).Handle)
		group.PATCH("/:uuid", handlers.NewUpdate(module.db, module.env).Handle)
		group.DELETE("/:uuid", handlers.NewDelete(module.db, module.env).Handle)
	}
}
