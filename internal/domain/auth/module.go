package auth

import (
	"auth/internal/domain/auth/handlers"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/rabbitmq"
	"auth/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	db             *gorm.DB
	redisClient    *goredis.Client
	publisher      *rabbitmq.Publisher
	userRepository userModel.Repository
	jwtService     *services.JWTService
	authConfig     config.AuthConfig
	env            config.Env
}

func NewModule(db *gorm.DB, redisClient *goredis.Client, publisher *rabbitmq.Publisher, authConfig config.AuthConfig, env config.Env) *Module {
	return &Module{
		db:             db,
		redisClient:    redisClient,
		publisher:      publisher,
		userRepository: userModel.NewRepository(db),
		jwtService:     services.NewJWTService(authConfig),
		authConfig:     authConfig,
		env:            env,
	}
}

func (module *Module) Register(group *gin.RouterGroup) {
	redisRepo := redis.NewRepository(module.redisClient)
	auth := group.Group("/auth")
	{
		auth.POST("/register", handlers.NewRegister(redisRepo, module.publisher, module.userRepository, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/login", handlers.NewLogin(redisRepo, module.userRepository, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/refresh", handlers.NewRefresh(redisRepo, module.userRepository, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/verify-email", handlers.NewVerifyEmail(module.userRepository, module.jwtService, module.env).Handle)
		auth.DELETE("/logout", handlers.NewLogout(redisRepo, module.env).Handle)
	}
}
