package middlewares

import (
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg config.AuthConfig, env config.Env) gin.HandlerFunc {
	jwtService := services.NewJWTService(cfg)

	return func(context *gin.Context) {
		authHeader := context.GetHeader("Authorization")

		if authHeader == "" {
			exceptions.NewUnauthorized(context, env).Throw(errors.New("authorization header is required"))
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if isInvalidStructure(parts) {
			exceptions.NewUnauthorized(context, env).Throw(errors.New("authorization header format must be Bearer {token}"))
			return
		}

		claims, err := jwtService.ValidateToken(parts[1], services.AccessTokenPurpose)

		if err != nil {
			exceptions.NewUnauthorized(context, env).Throw(errors.New("invalid or expired token"))
			return
		}

		context.Set("user_uuid", claims.UserUUID)
		context.Next()
	}
}

func isInvalidStructure(parts []string) bool {
	return len(parts) != 2 || parts[0] != "Bearer"
}
