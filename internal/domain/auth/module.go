package auth

import (
	"api/internal/domain/auth/handlers"
	"api/internal/domain/auth/services"
	userModel "api/internal/domain/user/model"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	db          *gorm.DB
	redisClient *goredis.Client
	userRepo    userModel.Repository
	jwtService  *services.JWTService
	authConfig  config.AuthConfig
	env         config.Env
}

func NewModule(db *gorm.DB, redisClient *goredis.Client, authConfig config.AuthConfig, env config.Env) *Module {
	return &Module{
		db:          db,
		redisClient: redisClient,
		userRepo:    userModel.NewRepository(db),
		jwtService:  services.NewJWTService(authConfig),
		authConfig:  authConfig,
		env:         env,
	}
}

func (module *Module) Register(group *gin.RouterGroup) {
	redisRepo := redis.NewRepository(module.redisClient)
	auth := group.Group("/auth")
	{
		auth.POST("/register", handlers.NewRegister(redisRepo, module.userRepo, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/login", handlers.NewLogin(redisRepo, module.userRepo, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/refresh", handlers.NewRefresh(redisRepo, module.userRepo, module.jwtService, module.authConfig, module.env).Handle)
		auth.DELETE("/logout", handlers.NewLogout(redisRepo, module.env).Handle)
	}
}
