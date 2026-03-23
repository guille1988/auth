package actions

import (
	"api/internal/infrastructure/redis"
	"context"
)

type Logout struct {
	redisRepository *redis.Repository
}

func NewLogout(redisRepository *redis.Repository) *Logout {
	return &Logout{redisRepository: redisRepository}
}

func (action *Logout) Execute(refreshToken string) error {
	return action.redisRepository.Delete(context.Background(), "auth:token:"+refreshToken)
}
