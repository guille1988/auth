package services

import (
	"auth/internal/infrastructure/config"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey               []byte
	accessTokenExpire       time.Duration
	emailVerificationExpire time.Duration
}

type TokenResponse struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
}

type TokenPurpose string

const (
	AccessTokenPurpose            TokenPurpose = "access"
	EmailVerificationTokenPurpose TokenPurpose = "email_verification"
)

type Claims struct {
	UserUUID string       `json:"user_uuid"`
	Purpose  TokenPurpose `json:"purpose"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg config.AuthConfig) *JWTService {
	return &JWTService{
		secretKey:               []byte(cfg.JWTSecret),
		accessTokenExpire:       cfg.AccessTokenExpire,
		emailVerificationExpire: cfg.EmailVerificationExpire,
	}
}

func (service *JWTService) GenerateEmailVerificationToken(userUUID string) (string, error) {
	expirationTime := time.Now().Add(service.emailVerificationExpire)

	claims := &Claims{
		UserUUID: userUUID,
		Purpose:  EmailVerificationTokenPurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(service.secretKey)
}

func (service *JWTService) GenerateAccessToken(userUUID string, refreshToken string) (*TokenResponse, error) {
	expirationTime := time.Now().Add(service.accessTokenExpire)

	claims := &Claims{
		UserUUID: userUUID,
		Purpose:  AccessTokenPurpose,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString(service.secretKey)

	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(service.accessTokenExpire.Seconds()),
	}, nil
}

func (service *JWTService) ValidateToken(tokenString string, expectedPurpose TokenPurpose) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)

		if !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return service.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Purpose != expectedPurpose {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
