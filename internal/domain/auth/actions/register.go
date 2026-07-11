package actions

import (
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/redis"
	"context"
	"log/slog"
	"time"

	"github.com/guille1988/go-app-shared/messaging/kafka/dtos"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type MessagePublisher interface {
	Publish(dto any) error
}

type Register struct {
	userRepository  userModel.Repository
	redisRepository *redis.Repository
	sessionIndex    *services.SessionIndex
	publisher       MessagePublisher
	jwtService      *services.JWTService
	authConfig      config.AuthConfig
}

func NewRegister(userRepository userModel.Repository, redisRepository *redis.Repository, publisher MessagePublisher, jwtService *services.JWTService, authConfig config.AuthConfig) *Register {
	return &Register{
		userRepository:  userRepository,
		redisRepository: redisRepository,
		sessionIndex:    services.NewSessionIndex(redisRepository),
		publisher:       publisher,
		jwtService:      jwtService,
		authConfig:      authConfig,
	}
}

func (action *Register) Execute(ctx context.Context, regData data.Register, device string) (*services.TokenResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(regData.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := userModel.User{
		UUID:     uuid.New(),
		Name:     regData.Name,
		Email:    regData.Email,
		Password: string(hashedPassword),
	}

	err = action.userRepository.Create(&user)

	if err != nil {
		return nil, err
	}

	var verificationToken string
	verificationToken, err = action.jwtService.GenerateEmailVerificationToken(user.UUID.String())

	if err != nil {
		return nil, err
	}

	verificationURL := action.authConfig.FrontendURL + "/verify-email?token=" + verificationToken

	err = action.publisher.Publish(dtos.WelcomeEmail{Email: user.Email, Name: user.Name, VerificationURL: verificationURL})

	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(action.authConfig.RefreshTokenExpire)

	sessionData := data.RefreshToken{
		UserUUID: user.UUID.String(),
		Device:   device,
	}

	err = action.redisRepository.Set(ctx, services.SessionKey(refreshToken), sessionData, time.Until(expiresAt))

	if err != nil {
		/*
			The account and verification email are already real and don't
			depend on Redis. Rather than fail the whole request (leaving the
			caller with a "phantom" account and no way to retry, since a
			second attempt would just hit "email already exists"), degrade:
			issue an access-only response with no refresh session. The user
			can log in normally once Redis recovers.
		*/
		slog.Error("failed to persist refresh session during register; issuing access-only response", "error", err)
		return action.jwtService.GenerateAccessToken(user.UUID.String(), "")
	}

	// Best-effort: see the same block in login.go.
	if indexErr := action.sessionIndex.Add(ctx, user.UUID.String(), refreshToken, expiresAt); indexErr != nil {
		slog.Error("failed to index refresh session", "error", indexErr)
	}

	return action.jwtService.GenerateAccessToken(user.UUID.String(), refreshToken)
}
