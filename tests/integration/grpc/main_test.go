package grpc

import (
	"auth/tests/integration"
	"testing"
)

func TestMain(test *testing.M) {
	integration.RunTests(test)
}
