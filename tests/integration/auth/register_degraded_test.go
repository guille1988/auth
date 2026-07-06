package auth

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/redis"
	"auth/tests/integration"
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRegisterDegradesWhenRedisFails(test *testing.T) {
	integration.TestCase(test, "it should still create the user and issue an access token when redis is unavailable", func(test *testing.T) {
		appInstance, err := integration.GetApp()
		assert.NoError(test, err)

		userRepository := userModel.NewRepository(appInstance.Container.DefaultConnection)
		jwtService := services.NewJWTService(appInstance.Config.Auth)

		/*
			A client pointed at a port nothing listens on, so Set() fails fast
			instead of hanging, simulating a real Redis outage.
		*/
		brokenRedisClient := goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:1"})
		defer func() { _ = brokenRedisClient.Close() }()
		brokenRedisRepository := redis.NewRepository(brokenRedisClient)

		registerAction := actions.NewRegister(userRepository, brokenRedisRepository, appInstance.Container.Publisher, jwtService, appInstance.Config.Auth)

		email := gofakeit.Email()

		registerData := data.Register{
			Name:     gofakeit.Name(),
			Email:    email,
			Password: "password123",
		}

		var response *services.TokenResponse
		response, err = registerAction.Execute(context.Background(), registerData, "test-agent")

		if assert.NoError(test, err, "register must not fail outright when only the redis session persist fails") {
			assert.NotEmpty(test, response.AccessToken, "an access token must still be issued")
			assert.Empty(test, response.RefreshToken, "no refresh token should be issued when the session couldn't be persisted")
		}

		user, findErr := userRepository.FindByEmail(email)
		assert.NoError(test, findErr, "the user must have been created despite the redis failure")
		assert.Equal(test, email, user.Email)
	})
}
