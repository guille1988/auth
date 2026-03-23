package exceptions

import (
	"auth/internal/infrastructure/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Exception represents an exception that can be thrown.
type Exception struct {
	Context    *gin.Context
	Env        config.Env
	StatusCode int
	Error      error
	Override   bool
}

// ExceptionResponse is the interface that all exceptions must implement.
type ExceptionResponse interface {
	Throw()
}

// newException creates a new exception instance.
func newException(context *gin.Context, env config.Env, statusCode int, override bool) *Exception {
	return &Exception{
		Context:    context,
		Env:        env,
		StatusCode: statusCode,
		Override:   override,
	}
}

// Throw throws the exception.
func (exception *Exception) Throw(err error) {
	var message string

	if err == nil || isWithoutPayload(exception) {
		message = http.StatusText(exception.StatusCode)
	} else {
		message = err.Error()
	}

	exception.Context.JSON(exception.StatusCode, gin.H{"error": message})
	exception.Context.Abort()
}

// isWithoutPayload returns true if the exception should not return a payload.
func isWithoutPayload(exception *Exception) bool {
	return exception.Env == config.ProductionEnv || !exception.Override
}
