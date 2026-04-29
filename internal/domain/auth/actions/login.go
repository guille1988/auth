package actions

import (
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/redis"
	"context"
	"errors"
	"time"

	"github.com/guille1988/go-app-shared/messaging/kafka/dtos"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailNotVerified = errors.New("email not verified")

type Login struct {
	userRepository  userModel.Repository
	redisRepository *redis.Repository
	jwtService      *services.JWTService
	authConfig      config.AuthConfig
	publisher       MessagePublisher
}

func NewLogin(userRepository userModel.Repository, redisRepository *redis.Repository, jwtService *services.JWTService, authConfig config.AuthConfig, publisher MessagePublisher) *Login {
	return &Login{
		userRepository:  userRepository,
		redisRepository: redisRepository,
		jwtService:      jwtService,
		authConfig:      authConfig,
		publisher:       publisher,
	}
}

func (action *Login) Execute(loginData data.Login, device string) (*services.TokenResponse, error) {
	user, err := action.userRepository.FindByEmail(loginData.Email)

	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))

	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if user.EmailVerifiedAt == nil {
		return nil, ErrEmailNotVerified
	}

	now := time.Now()
	_ = action.userRepository.Update(user, map[string]any{"last_login_at": now})

	refreshToken := uuid.New().String()
	expiresAt := now.Add(action.authConfig.RefreshTokenExpire)

	sessionData := data.RefreshToken{
		UserID: user.ID,
		Device: device,
	}

	err = action.redisRepository.Set(context.Background(), "auth:token:"+refreshToken, sessionData, time.Until(expiresAt))

	if err != nil {
		return nil, err
	}

	err = action.publisher.Publish(dtos.UserLoggedIn{Email: user.Email, Name: user.Name})

	if err != nil {
		return nil, err
	}

	return action.jwtService.GenerateAccessToken(user.UUID.String(), refreshToken)
}
