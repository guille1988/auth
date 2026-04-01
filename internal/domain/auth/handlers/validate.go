package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ValidateHandler struct{}

func NewValidate() *ValidateHandler {
	return &ValidateHandler{}
}

func (handler *ValidateHandler) Handle(ctx *gin.Context) {
	ctx.Header("X-User-UUID", ctx.GetString("user_uuid"))
	ctx.Status(http.StatusOK)
}
