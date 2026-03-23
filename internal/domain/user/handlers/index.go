package handlers

import (
	"auth/internal/domain/user/model"
	"auth/internal/domain/user/responses"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type IndexHandler struct {
	repository model.Repository
	env        config.Env
}

func NewIndex(db *gorm.DB, env config.Env) *IndexHandler {
	repository := model.NewRepository(db)
	return &IndexHandler{
		repository: repository,
		env:        env,
	}
}

func (handler *IndexHandler) Handle(context *gin.Context) {
	users, err := handler.repository.FindAll()

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	responses.NewIndexResponse(context).Make(users)
}
