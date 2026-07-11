package grpc

import (
	"auth/internal/domain/auth/services"
	"context"
	"errors"

	authv1 "github.com/guille1988/go-app-shared/rpc/auth/v1"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TokenValidator is the subset of JWTService this transport needs.
type TokenValidator interface {
	ValidateToken(tokenString string, expectedPurpose services.TokenPurpose) (*services.Claims, error)
}

// SessionChecker is the subset of SessionIndex this transport needs.
type SessionChecker interface {
	HasLiveSession(ctx context.Context, userUUID string) (bool, error)
}

/*
Server is the gRPC transport adapter for the auth domain, the gRPC
analogue of the gin handlers.
*/
type Server struct {
	authv1.UnimplementedAuthServiceServer
	validator TokenValidator
	sessions  SessionChecker
}

func NewServer(validator TokenValidator, sessions SessionChecker) *Server {
	return &Server{validator: validator, sessions: sessions}
}

/*
ValidateToken reports whether an access token is still valid: the JWT must
verify, AND its user must hold at least one live session (a logged-out user
gets REVOKED). An invalid token is a domain result (valid=false plus a
reason), never a gRPC error: transport errors are reserved for
infrastructure failures — a session-store outage in particular returns
Unavailable so callers fail to open instead of revoking everyone.
*/
func (server *Server) ValidateToken(ctx context.Context, request *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	if request.GetToken() == "" {
		return invalid(authv1.ValidateTokenResponse_MALFORMED), nil
	}

	claims, err := server.validator.ValidateToken(request.GetToken(), services.AccessTokenPurpose)

	switch {
	case err == nil:
		live, sessionErr := server.sessions.HasLiveSession(ctx, claims.UserUUID)

		if sessionErr != nil {
			return nil, status.Error(codes.Unavailable, "session store unreachable")
		}

		if !live {
			return invalid(authv1.ValidateTokenResponse_REVOKED), nil
		}

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
