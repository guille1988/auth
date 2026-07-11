package services

import (
	"auth/internal/domain/auth/services"
	"auth/internal/infrastructure/redis"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestSessionIndex() (*services.SessionIndex, func()) {
	client := goredis.NewClient(&goredis.Options{Addr: "redis:6379", Password: "auth"})
	return services.NewSessionIndex(redis.NewRepository(client)), func() { _ = client.Close() }
}

func TestSessionIndex(test *testing.T) {
	index, closeClient := newTestSessionIndex()
	defer closeClient()

	ctx := context.Background()

	test.Run("it should report a live session after adding one", func(test *testing.T) {
		userUUID := uuid.New().String()

		assert.NoError(test, index.Add(ctx, userUUID, uuid.New().String(), time.Now().Add(time.Hour)))

		live, err := index.HasLiveSession(ctx, userUUID)
		assert.NoError(test, err)
		assert.True(test, live)
	})

	test.Run("it should report no live session after removing the only one", func(test *testing.T) {
		userUUID := uuid.New().String()
		refreshToken := uuid.New().String()

		assert.NoError(test, index.Add(ctx, userUUID, refreshToken, time.Now().Add(time.Hour)))
		assert.NoError(test, index.Remove(ctx, userUUID, refreshToken))

		live, err := index.HasLiveSession(ctx, userUUID)
		assert.NoError(test, err)
		assert.False(test, live)
	})

	test.Run("it should purge sessions whose expiry already passed", func(test *testing.T) {
		userUUID := uuid.New().String()

		assert.NoError(test, index.Add(ctx, userUUID, uuid.New().String(), time.Now().Add(-time.Minute)))

		live, err := index.HasLiveSession(ctx, userUUID)
		assert.NoError(test, err)
		assert.False(test, live)
	})

	test.Run("it should keep the user live while at least one of several sessions remains", func(test *testing.T) {
		userUUID := uuid.New().String()
		firstSession := uuid.New().String()
		secondSession := uuid.New().String()

		assert.NoError(test, index.Add(ctx, userUUID, firstSession, time.Now().Add(time.Hour)))
		assert.NoError(test, index.Add(ctx, userUUID, secondSession, time.Now().Add(time.Hour)))
		assert.NoError(test, index.Remove(ctx, userUUID, firstSession))

		live, err := index.HasLiveSession(ctx, userUUID)
		assert.NoError(test, err)
		assert.True(test, live)
	})

	test.Run("it should report no live session for an unknown user", func(test *testing.T) {
		live, err := index.HasLiveSession(ctx, uuid.New().String())
		assert.NoError(test, err)
		assert.False(test, live)
	})
}
