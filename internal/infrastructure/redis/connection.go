package redis

import (
	"auth/internal/infrastructure/config"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewConnection(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.Database,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("could not connect to redis: %w (also failed to close client: %v)", err, closeErr)
		}

		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}

	return client, nil
}
