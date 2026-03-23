package handlers

import (
	"api/internal/domain/user/actions"
	"api/internal/domain/user/model"
	"api/internal/domain/user/responses"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/exceptions"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ShowHandler struct {
	showUser actions.Show
	env      config.Env
}

func NewShow(db *gorm.DB, env config.Env) *ShowHandler {
	repository := model.NewRepository(db)
	showUser := actions.NewShow(repository)

	return &ShowHandler{
		showUser: showUser,
		env:      env,
	}
}

func (handler *ShowHandler) Handle(context *gin.Context) {
	uuid := context.Param("uuid")

	user, err := handler.showUser.Execute(uuid)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			exceptions.NewNotFound(context, handler.env).Throw(err)
			return
		}

		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	responses.NewShowResponse(context).Make(user)
}
