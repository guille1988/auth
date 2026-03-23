package actions

import (
	"api/internal/domain/auth/data"
	"api/internal/domain/auth/services"
	userModel "api/internal/domain/user/model"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/redis"
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

func (action *Refresh) Execute(refreshToken string) (*services.TokenResponse, error) {
	ctx := context.Background()
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
