package container

import (
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/database"
	"api/internal/infrastructure/redis"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	DefaultConnection *gorm.DB
	Redis             *goredis.Client
}

// New creates a new container with initialized database connections.
func New(dbCfg config.DatabaseConfig, redisCfg config.RedisConfig) (*Container, error) {
	defaultConnection, err := database.NewConnection(dbCfg.Connections[config.Default])

	if err != nil {
		return nil, err
	}

	var redisClient *goredis.Client
	redisClient, err = redis.NewConnection(redisCfg)

	if err != nil {
		return nil, err
	}

	return &Container{
		DefaultConnection: defaultConnection,
		Redis:             redisClient,
	}, nil
}
