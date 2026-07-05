package actions

import (
	"auth/internal/infrastructure/redis"
	"context"
)

type Logout struct {
	redisRepository *redis.Repository
}

func NewLogout(redisRepository *redis.Repository) *Logout {
	return &Logout{redisRepository: redisRepository}
}

func (action *Logout) Execute(ctx context.Context, refreshToken string) error {
	return action.redisRepository.Delete(ctx, "auth:token:"+refreshToken)
}
