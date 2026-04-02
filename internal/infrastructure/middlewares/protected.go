package middlewares

import (
	"auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

func ProtectedGroup(group *gin.RouterGroup, cfg config.AuthConfig, repo model.Repository, env config.Env) *gin.RouterGroup {
	return group.Group("", AuthMiddleware(cfg, env), EnsureEmailVerified(repo, env))
}
