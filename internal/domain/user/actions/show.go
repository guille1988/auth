package actions

import (
	"api/internal/domain/user/model"
)

type Show struct {
	repository model.Repository
}

func NewShow(repository model.Repository) Show {
	return Show{repository: repository}
}

func (show Show) Execute(uuid string) (*model.User, error) {
	return show.repository.FindByUUID(uuid)
}
