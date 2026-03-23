package actions

import (
	"api/internal/domain/auth/data"
	"api/internal/domain/auth/services"
	userModel "api/internal/domain/user/model"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/redis"
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Register struct {
	userRepository  userModel.Repository
	redisRepository *redis.Repository
	jwtService      *services.JWTService
	authConfig      config.AuthConfig
}

func NewRegister(userRepository userModel.Repository, redisRepository *redis.Repository, jwtService *services.JWTService, authConfig config.AuthConfig) *Register {
	return &Register{
		userRepository:  userRepository,
		redisRepository: redisRepository,
		jwtService:      jwtService,
		authConfig:      authConfig,
	}
}

func (action *Register) Execute(regData data.Register, device string) (*services.TokenResponse, error) {
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

	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(action.authConfig.RefreshTokenExpire)

	sessionData := data.RefreshToken{
		UserID: user.ID,
		Device: device,
	}

	err = action.redisRepository.Set(context.Background(), "auth:token:"+refreshToken, sessionData, time.Until(expiresAt))

	if err != nil {
		return nil, err
	}

	return action.jwtService.GenerateAccessToken(user.UUID.String(), refreshToken)
}
