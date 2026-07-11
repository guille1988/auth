package actions

import (
	"auth/internal/domain/auth/data"
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/redis"
	"context"
	"errors"
	"log/slog"

	goredis "github.com/redis/go-redis/v9"
)

type Logout struct {
	redisRepository *redis.Repository
	sessionIndex    *services.SessionIndex
}

func NewLogout(redisRepository *redis.Repository) *Logout {
	return &Logout{
		redisRepository: redisRepository,
		sessionIndex:    services.NewSessionIndex(redisRepository),
	}
}

/*
Execute deletes the session and removes it from the per-user index. GetDel
(not Delete) recovers the session value atomically, which carries the user
UUID the index is keyed by. An unknown token stays an idempotent success,
and index maintenance is best-effort: once the session key is gone, the
logout succeeded.
*/
func (action *Logout) Execute(ctx context.Context, refreshToken string) error {
	var sessionData data.RefreshToken
	err := action.redisRepository.GetDel(ctx, services.SessionKey(refreshToken), &sessionData)

	if errors.Is(err, goredis.Nil) {
		return nil
	}

	if err != nil {
		return err
	}

	if sessionData.UserUUID == "" {
		return nil
	}

	if removeErr := action.sessionIndex.Remove(ctx, sessionData.UserUUID, refreshToken); removeErr != nil {
		slog.Warn("logout: failed to remove session index entry", "error", removeErr)
	}

	return nil
}
