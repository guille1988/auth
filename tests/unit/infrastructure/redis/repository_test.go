package redis

import (
	"auth/internal/infrastructure/redis"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestRepository() (*redis.Repository, func()) {
	client := goredis.NewClient(&goredis.Options{Addr: "redis:6379", Password: "auth"})
	return redis.NewRepository(client), func() { _ = client.Close() }
}

func TestSortedSetOperations(test *testing.T) {
	repository, closeClient := newTestRepository()
	defer closeClient()

	ctx := context.Background()

	test.Run("it should purge members with expired scores when counting live ones", func(test *testing.T) {
		key := "test:sorted:" + uuid.New().String()
		now := time.Now()

		assert.NoError(test, repository.AddToSortedSet(ctx, key, "expired-member", float64(now.Add(-time.Minute).Unix()), time.Hour))
		assert.NoError(test, repository.AddToSortedSet(ctx, key, "live-member", float64(now.Add(time.Hour).Unix()), time.Hour))

		count, err := repository.CountLiveMembers(ctx, key, now)
		assert.NoError(test, err)
		assert.Equal(test, int64(1), count, "the expired member must be purged, the live one counted")
	})

	test.Run("it should count zero for a missing key", func(test *testing.T) {
		count, err := repository.CountLiveMembers(ctx, "test:sorted:"+uuid.New().String(), time.Now())
		assert.NoError(test, err)
		assert.Equal(test, int64(0), count)
	})

	test.Run("it should remove members without failing on absent ones", func(test *testing.T) {
		key := "test:sorted:" + uuid.New().String()

		assert.NoError(test, repository.AddToSortedSet(ctx, key, "member", float64(time.Now().Add(time.Hour).Unix()), time.Hour))
		assert.NoError(test, repository.RemoveFromSortedSet(ctx, key, "member"))
		assert.NoError(test, repository.RemoveFromSortedSet(ctx, key, "never-existed"))

		count, err := repository.CountLiveMembers(ctx, key, time.Now())
		assert.NoError(test, err)
		assert.Equal(test, int64(0), count)
	})
}
