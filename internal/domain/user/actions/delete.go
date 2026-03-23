package actions

import (
	"api/internal/domain/user/model"
)

type Delete struct {
	repository model.Repository
}

func NewDelete(repository model.Repository) Delete {
	return Delete{repository: repository}
}

func (delete Delete) Execute(user *model.User) error {
	return delete.repository.Delete(user)
}
