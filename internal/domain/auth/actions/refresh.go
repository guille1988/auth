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

	"github.com/google/uuid"
)

type Refresh struct {
	userRepository  userModel.Repository
	redisRepository *redis.Repository
	jwtService      *services.JWTService
	authConfig      config.AuthConfig
}

func NewRefresh(userRepository userModel.Repository, redisRepository *redis.Repository, jwtService *services.JWTService, authConfig config.AuthConfig) *Refresh {
	return &Refresh{
		userRepository:  userRepository,
		redisRepository: redisRepository,
		jwtService:      jwtService,
		authConfig:      authConfig,
	}
}

func (action *Refresh) Execute(ctx context.Context, refreshToken string) (*services.TokenResponse, error) {
	var sessionData data.RefreshToken
	tokenKey := "auth:token:" + refreshToken
	err := action.redisRepository.Get(ctx, tokenKey, &sessionData)

	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	var user *userModel.User
	user, err = action.userRepository.FindByID(sessionData.UserID)

	if err != nil {
		return nil, errors.New("user not found")
	}

	_ = action.redisRepository.Delete(ctx, tokenKey)

	newRefreshToken := uuid.New().String()
	expiresAt := time.Now().Add(action.authConfig.RefreshTokenExpire)

	err = action.redisRepository.Set(ctx, "auth:token:"+newRefreshToken, sessionData, time.Until(expiresAt))

	if err != nil {
		return nil, err
	}

	return action.jwtService.GenerateAccessToken(user.UUID.String(), newRefreshToken)
}
