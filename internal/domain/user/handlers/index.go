package handlers

import (
	"api/internal/domain/user/model"
	"api/internal/domain/user/responses"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/exceptions"

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
