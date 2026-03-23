package handlers

import (
	"auth/internal/domain/user/actions"
	"auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteHandler struct {
	showUser   actions.Show
	deleteUser actions.Delete
	env        config.Env
}

func NewDelete(db *gorm.DB, env config.Env) *DeleteHandler {
	repository := model.NewRepository(db)
	showUser := actions.NewShow(repository)
	deleteUser := actions.NewDelete(repository)

	return &DeleteHandler{
		showUser:   showUser,
		deleteUser: deleteUser,
		env:        env,
	}
}

func (handler *DeleteHandler) Handle(context *gin.Context) {
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

	err = handler.deleteUser.Execute(user)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	context.Status(http.StatusNoContent)
}
