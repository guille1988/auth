package auth

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/domain/auth/handlers"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/middlewares"
	"auth/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	db             *gorm.DB
	redisClient    *goredis.Client
	publisher      actions.MessagePublisher
	userRepository userModel.Repository
	jwtService     *services.JWTService
	authConfig     config.AuthConfig
	env            config.Env
}

func NewModule(db *gorm.DB, redisClient *goredis.Client, publisher actions.MessagePublisher, authConfig config.AuthConfig, env config.Env) *Module {
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
		protected := middlewares.ProtectedGroup(auth, module.authConfig, module.userRepository, module.env)
		{
			protected.GET("/validate", handlers.NewValidate().Handle)
		}

		auth.POST("/register", handlers.NewRegister(redisRepo, module.publisher, module.userRepository, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/login", handlers.NewLogin(redisRepo, module.publisher, module.userRepository, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/refresh", handlers.NewRefresh(redisRepo, module.userRepository, module.jwtService, module.authConfig, module.env).Handle)
		auth.POST("/verify-email", handlers.NewVerifyEmail(module.userRepository, module.jwtService, module.env).Handle)
		auth.POST("/resend-verification", handlers.NewResendVerificationEmail(module.userRepository, module.jwtService, module.publisher, module.authConfig, module.env).Handle)
		auth.DELETE("/logout", handlers.NewLogout(redisRepo, module.env).Handle)
	}
}
