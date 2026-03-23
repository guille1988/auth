package actions

import (
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"errors"
	"time"
)

type VerifyEmail struct {
	userRepository userModel.Repository
	jwtService     *services.JWTService
}

func NewVerifyEmail(userRepository userModel.Repository, jwtService *services.JWTService) *VerifyEmail {
	return &VerifyEmail{
		userRepository: userRepository,
		jwtService:     jwtService,
	}
}

func (action *VerifyEmail) Execute(token string) error {
	claims, err := action.jwtService.ValidateToken(token)

	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	var user *userModel.User
	user, err = action.userRepository.FindByUUID(claims.UserUUID)

	if err != nil {
		return err
	}

	if user.EmailVerifiedAt != nil {
		return nil
	}

	now := time.Now()
	return action.userRepository.Update(user, map[string]any{
		"email_verified_at": &now,
	})
}
