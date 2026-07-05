package redis

import (
	"auth/internal/infrastructure/config"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConnectionReturnsUnderlyingErrorOnPingFailure(test *testing.T) {
	cfg := config.RedisConfig{
		Host:     "127.0.0.1",
		Port:     "1", // nothing listens here, so Ping must fail fast
		Password: "",
		Database: 0,
	}

	client, err := NewConnection(cfg)

	assert.Nil(test, client)
	assert.Error(test, err)
	assert.Contains(test, err.Error(), "could not connect to redis")
	assert.NotContains(test, err.Error(), "%!w(<nil>)", "the original ping error must not be swallowed by the Close() error")
	assert.False(test, strings.HasSuffix(err.Error(), ": "), "the wrapped error message must not be empty")
}
