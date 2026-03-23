package handlers

import (
	"api/internal/domain/user/actions"
	"api/internal/domain/user/data"
	"api/internal/domain/user/model"
	"api/internal/infrastructure/config"
	"api/internal/infrastructure/exceptions"
	"api/internal/infrastructure/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StoreHandler struct {
	storeUser actions.Store
	env       config.Env
}

func NewStore(db *gorm.DB, env config.Env) *StoreHandler {
	repository := model.NewRepository(db)
	storeUser := actions.NewStore(repository)

	return &StoreHandler{
		storeUser: storeUser,
		env:       env,
	}
}

func (handler *StoreHandler) Handle(context *gin.Context) {
	var storeUserData data.StoreUser
	if validator.New(context, handler.env).Fails(&storeUserData) {
		return
	}

	err := handler.storeUser.Execute(storeUserData)

	if err != nil {
		exceptions.NewServer(context, handler.env).Throw(err)
		return
	}

	context.Status(http.StatusCreated)
}
