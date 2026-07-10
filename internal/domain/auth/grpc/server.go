package grpc

import (
	"auth/internal/domain/auth/services"
	"context"
	"errors"

	authv1 "github.com/guille1988/go-app-shared/rpc/auth/v1"

	"github.com/golang-jwt/jwt/v5"
)

// TokenValidator is the subset of JWTService this transport needs.
type TokenValidator interface {
	ValidateToken(tokenString string, expectedPurpose services.TokenPurpose) (*services.Claims, error)
}

/*
Server is the gRPC transport adapter for the auth domain, the gRPC
analogue of the gin handlers.
*/
type Server struct {
	authv1.UnimplementedAuthServiceServer
	validator TokenValidator
}

func NewServer(validator TokenValidator) *Server {
	return &Server{validator: validator}
}

/*
ValidateToken reports whether an access token is still valid. An invalid
token is a domain result (valid=false plus a reason), never a gRPC error:
transport errors are reserved for infrastructure failures, so callers can
tell "invalid token" apart from "auth is unavailable".
*/
func (server *Server) ValidateToken(_ context.Context, request *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if request.GetToken() == "" {
		return invalid(authv1.ValidateTokenResponse_MALFORMED), nil
	}

	claims, err := server.validator.ValidateToken(request.GetToken(), services.AccessTokenPurpose)

	switch {
	case err == nil:
		return &authv1.ValidateTokenResponse{Valid: true, UserUuid: claims.UserUUID}, nil
	case errors.Is(err, jwt.ErrTokenExpired):
		return invalid(authv1.ValidateTokenResponse_EXPIRED), nil
	case errors.Is(err, services.ErrWrongPurpose):
		return invalid(authv1.ValidateTokenResponse_WRONG_PURPOSE), nil
	default:
		return invalid(authv1.ValidateTokenResponse_MALFORMED), nil
	}
}

func invalid(reason authv1.ValidateTokenResponse_Reason) *authv1.ValidateTokenResponse {
	return &authv1.ValidateTokenResponse{Valid: false, Reason: reason}
}
