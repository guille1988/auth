package container

import (
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/database"
	"auth/internal/infrastructure/rabbitmq"
	"auth/internal/infrastructure/redis"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	DefaultConnection *gorm.DB
	Redis             *goredis.Client
	Publisher         *rabbitmq.Publisher
}

// New creates a new container with initialized database connections.
func New(dbCfg config.DatabaseConfig, redisCfg config.RedisConfig, rabbitCfg config.RabbitMQConfig) (*Container, error) {
	defaultConnection, err := database.NewConnection(dbCfg.Connections[config.Default])

	if err != nil {
		return nil, err
	}

	var redisClient *goredis.Client
	redisClient, err = redis.NewConnection(redisCfg)

	if err != nil {
		return nil, err
	}

	var publisher *rabbitmq.Publisher
	publisher, err = rabbitmq.NewPublisher(rabbitCfg)

	if err != nil {
		_ = redisClient.Close()

		sqlDB, _ := defaultConnection.DB()
		_ = sqlDB.Close()

		return nil, err
	}

	return &Container{
		DefaultConnection: defaultConnection,
		Redis:             redisClient,
		Publisher:         publisher,
	}, nil
}
