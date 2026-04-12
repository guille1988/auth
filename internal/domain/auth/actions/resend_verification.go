package actions

import (
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"context"

	"github.com/guille1988/go-app-shared/messaging/kafka/dtos"
)

type ResendVerificationEmail struct {
	userRepository userModel.Repository
	jwtService     *services.JWTService
	publisher      MessagePublisher
	authConfig     config.AuthConfig
}

func NewResendVerificationEmail(userRepository userModel.Repository, jwtService *services.JWTService, publisher MessagePublisher, authConfig config.AuthConfig) *ResendVerificationEmail {
	return &ResendVerificationEmail{
		userRepository: userRepository,
		jwtService:     jwtService,
		publisher:      publisher,
		authConfig:     authConfig,
	}
}

func (action *ResendVerificationEmail) Execute(ctx context.Context, email string) error {
	user, err := action.userRepository.FindByEmail(email)

	if err != nil || user.EmailVerifiedAt != nil {
		return nil
	}

	var verificationToken string
	verificationToken, err = action.jwtService.GenerateEmailVerificationToken(user.UUID.String())

	if err != nil {
		return err
	}

	verificationURL := action.authConfig.FrontendURL + "/verify-email?token=" + verificationToken

	return action.publisher.Publish(dtos.WelcomeEmail{Email: user.Email, Name: user.Name, VerificationURL: verificationURL})
}
