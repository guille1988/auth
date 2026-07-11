package actions

import (
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/redis"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Refresh struct {
	userRepository  userModel.Repository
	redisRepository *redis.Repository
	sessionIndex    *services.SessionIndex
	jwtService      *services.JWTService
	authConfig      config.AuthConfig
}

func NewRefresh(userRepository userModel.Repository, redisRepository *redis.Repository, jwtService *services.JWTService, authConfig config.AuthConfig) *Refresh {
	return &Refresh{
		userRepository:  userRepository,
		redisRepository: redisRepository,
		sessionIndex:    services.NewSessionIndex(redisRepository),
		jwtService:      jwtService,
		authConfig:      authConfig,
	}
}

func (action *Refresh) Execute(ctx context.Context, refreshToken string) (*services.TokenResponse, error) {
	var sessionData data.RefreshToken
	tokenKey := services.SessionKey(refreshToken)
	err := action.redisRepository.GetDel(ctx, tokenKey, &sessionData)

	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	var user *userModel.User
	user, err = action.userRepository.FindByUUID(sessionData.UserUUID)

	if err != nil {
		return nil, errors.New("user not found")
	}

	newRefreshToken := uuid.New().String()
	expiresAt := time.Now().Add(action.authConfig.RefreshTokenExpire)

	err = action.redisRepository.Set(ctx, services.SessionKey(newRefreshToken), sessionData, time.Until(expiresAt))

	if err != nil {
		return nil, err
	}

	/*
	 Best-effort index rotation, Add before Remove: a single-session user
	 must never hit a zero-member window mid-rotation, or a revalidation
	 tick landing in between would spuriously revoke their connections.
	*/
	if indexErr := action.sessionIndex.Add(ctx, user.UUID.String(), newRefreshToken, expiresAt); indexErr != nil {
		slog.Error("failed to index rotated refresh session", "error", indexErr)
	}

	if indexErr := action.sessionIndex.Remove(ctx, user.UUID.String(), refreshToken); indexErr != nil {
		slog.Error("failed to remove rotated session from index", "error", indexErr)
	}

	return action.jwtService.GenerateAccessToken(user.UUID.String(), newRefreshToken)
}
