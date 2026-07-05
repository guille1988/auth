package actions_test

import (
	"auth/internal/domain/auth/actions"
	"auth/internal/infrastructure/redis"
	"context"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

/*
These tests prove the caller's context is actually wired through to the
underlying Redis call, rather than a hardcoded context.Background() that
silently ignores cancellation. An already-canceled context must cause the
Redis operation to fail fast with context.Canceled.
*/
func newTestRedisRepository() (*redis.Repository, func()) {
	client := goredis.NewClient(&goredis.Options{Addr: "redis:6379", Password: "auth"})
	return redis.NewRepository(client), func() { _ = client.Close() }
}

func TestLogoutExecutePropagatesContextCancellation(test *testing.T) {
	repository, closeClient := newTestRedisRepository()
	defer closeClient()

	logoutAction := actions.NewLogout(repository)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := logoutAction.Execute(ctx, "some-refresh-token")

	assert.ErrorIs(test, err, context.Canceled)
}
