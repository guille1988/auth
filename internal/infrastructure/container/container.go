package container

import (
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/database"
	"auth/internal/infrastructure/providers/messaging"
	"auth/internal/infrastructure/rabbitmq"
	"auth/internal/infrastructure/redis"
	"auth/internal/shared/messaging/rabbitmq/dtos"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
	DefaultConnection *gorm.DB
	Redis             *goredis.Client
	RabbitMQProvider  *messaging.RabbitMQRegister
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
		sqlDB, _ := defaultConnection.DB()
		_ = sqlDB.Close()

		return nil, err
	}

	return &Container{
		DefaultConnection: defaultConnection,
		Redis:             redisClient,
	}, nil
}

// InitPublisher initializes the RabbitMQ publisher and registers known routes.
func (container *Container) InitPublisher(rabbitCfg config.RabbitMQConfig) error {
	publisher, err := rabbitmq.NewPublisher(rabbitCfg)

	if err != nil {
		return err
	}

	provider := messaging.NewRabbitMQRegister(publisher)

	err = provider.Register(dtos.WelcomeEmail{}, messaging.Route{
		Exchange:     "auth.events",
		RoutingKey:   "user.created",
		ExchangeType: "topic",
	})

	if err != nil {
		_ = publisher.Close()
		return err
	}

	container.RabbitMQProvider = provider

	return nil
}
