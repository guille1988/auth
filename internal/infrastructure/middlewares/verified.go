package middlewares

import (
	"auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"errors"

	"github.com/gin-gonic/gin"
)

func EnsureEmailVerified(userRepository model.Repository, env config.Env) gin.HandlerFunc {
	return func(context *gin.Context) {
		userUUID := context.GetString("user_uuid")

		user, err := userRepository.FindByUUID(userUUID)

		if isNotVerified(err, user) {
			exceptions.NewForbidden(context, env).Throw(errors.New("email not verified"))
			return
		}

		context.Next()
	}
}

func isNotVerified(err error, user *model.User) bool {
	return err != nil || user.EmailVerifiedAt == nil
}
