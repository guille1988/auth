package actions

import (
	"api/internal/domain/user/data"
	"api/internal/domain/user/model"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Store struct {
	repository model.Repository
}

func NewStore(repository model.Repository) Store {
	return Store{repository: repository}
}

func (store Store) Execute(data data.StoreUser) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user := model.User{
		UUID:     uuid.New(),
		Name:     data.Name,
		Email:    data.Email,
		Password: string(hashedPassword),
	}

	return store.repository.Create(&user)
}
