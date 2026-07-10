package config

import (
	"auth/internal/infrastructure/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRejectsInsecureDefaultJWTSecretInProduction(test *testing.T) {
	test.Setenv("APP_ENV", "production")
	test.Setenv("AUTH_JWT_SECRET", "secret")

	_, err := config.New()

	if assert.Error(test, err) {
		assert.Contains(test, err.Error(), "AUTH_JWT_SECRET")
	}
}

func TestNewRejectsEmptyJWTSecretInProduction(test *testing.T) {
	test.Setenv("APP_ENV", "production")
	test.Setenv("AUTH_JWT_SECRET", "")

	_, err := config.New()

	assert.Error(test, err)
}

func TestNewAllowsRealJWTSecretInProduction(test *testing.T) {
	test.Setenv("APP_ENV", "production")
	test.Setenv("AUTH_JWT_SECRET", "a-real-strong-secret-value")

	_, err := config.New()

	assert.NoError(test, err)
}

func TestNewAllowsInsecureDefaultOutsideProduction(test *testing.T) {
	test.Setenv("APP_ENV", "local")
	test.Setenv("AUTH_JWT_SECRET", "secret")

	_, err := config.New()

	assert.NoError(test, err, "the insecure default must still work for local/dev")
}
