package handlers

import (
	"auth/internal/domain/user/actions"
	"auth/internal/domain/user/data"
	"auth/internal/domain/user/model"
	"auth/internal/infrastructure/config"
	"auth/internal/infrastructure/exceptions"
	"auth/internal/infrastructure/validator"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateHandler struct {
	showUser   actions.Show
	updateUser actions.Update
	env        config.Env
}

func NewUpdate(db *gorm.DB, env config.Env) *UpdateHandler {
	repository := model.NewRepository(db)
	showUser := actions.NewShow(repository)
	updateUser := actions.NewUpdate(repository)

	return &UpdateHandler{
		showUser:   showUser,
		updateUser: updateUser,
		env:        env,
	}
}

func (handler *UpdateHandler) Handle(context *gin.Context) {
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

	var updateUserData data.UpdateUser
	if validator.New(context, handler.env).Fails(&updateUserData) {
		return
	}

	err = handler.updateUser.Execute(user, updateUserData)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	context.Status(http.StatusNoContent)
}
