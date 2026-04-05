package actions

import (
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/redis"
	"context"
	"time"

	"github.com/guille1988/go-app-shared/messaging/rabbitmq/dtos"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type MessagePublisher interface {
	Publish(ctx context.Context, dto any) error
}

type Register struct {
	userRepository  userModel.Repository
	redisRepository *redis.Repository
	publisher       MessagePublisher
	jwtService      *services.JWTService
	authConfig      config.AuthConfig
}

func NewRegister(userRepository userModel.Repository, redisRepository *redis.Repository, publisher MessagePublisher, jwtService *services.JWTService, authConfig config.AuthConfig) *Register {
	return &Register{
		userRepository:  userRepository,
		redisRepository: redisRepository,
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

	err = action.publisher.Publish(ctx, dtos.WelcomeEmail{Email: user.Email, Name: user.Name, VerificationURL: verificationURL})

	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(action.authConfig.RefreshTokenExpire)

	sessionData := data.RefreshToken{
		UserID: user.ID,
		Device: device,
	}

	err = action.redisRepository.Set(ctx, "auth:token:"+refreshToken, sessionData, time.Until(expiresAt))

	if err != nil {
		return nil, err
	}

	return action.jwtService.GenerateAccessToken(user.UUID.String(), refreshToken)
}
