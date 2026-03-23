package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	client *redis.Client
}

func NewRepository(client *redis.Client) *Repository {
	return &Repository{client: client}
}

func (repository *Repository) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)

	if err != nil {
		return err
	}

	return repository.client.Set(ctx, key, data, expiration).Err()
}

func (repository *Repository) Get(ctx context.Context, key string, dest any) error {
	data, err := repository.client.Get(ctx, key).Bytes()

	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

func (repository *Repository) Delete(ctx context.Context, key string) error {
	return repository.client.Del(ctx, key).Err()
}
