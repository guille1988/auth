package validator

import (
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/exceptions"

	"github.com/gin-gonic/gin"
)

type Validator struct {
	context *gin.Context
	Env     config.Env
}

type ValidationResponse struct {
	Errors []string `json:"errors,omitzero"`
}

func New(context *gin.Context, env config.Env) *Validator {
	return &Validator{
		context: context,
		Env:     env,
	}
}

// Passes checks if the data passes the validation.
func (validator *Validator) Passes(data any) bool {
	err := validator.context.ShouldBindJSON(data)

	if err != nil {
		exceptions.NewUnprocessableEntity(validator.context, validator.Env).Throw(err)
		return false
	}

	return true
}

// Fails checks if the data fails the validation.
func (validator *Validator) Fails(data any) bool {
	return validator.Passes(data) == false
}
