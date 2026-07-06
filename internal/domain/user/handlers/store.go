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

type StoreHandler struct {
	repository model.Repository
	storeUser  actions.Store
	env        config.Env
}

func NewStore(db *gorm.DB, env config.Env) *StoreHandler {
	repository := model.NewRepository(db)
	storeUser := actions.NewStore(repository)

	return &StoreHandler{
		repository: repository,
		storeUser:  storeUser,
		env:        env,
	}
}

func (handler *StoreHandler) Handle(context *gin.Context) {
	var storeUserData data.StoreUser
	if validator.New(context, handler.env).Fails(&storeUserData) {
		return
	}

	exists, err := handler.repository.ExistByEmail(storeUserData.Email)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	if exists {
		exceptions.NewUnprocessableEntity(context, handler.env).
			Throw(errors.New("this e-mail already exists"))

		return
	}

	err = handler.storeUser.Execute(storeUserData)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	context.Status(http.StatusCreated)
}
