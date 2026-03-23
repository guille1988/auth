package actions

import (
	"api/internal/domain/user/data"
	"api/internal/domain/user/model"
)

type Update struct {
	repository model.Repository
}

func NewUpdate(repository model.Repository) Update {
	return Update{repository: repository}
}

func (update Update) Execute(user *model.User, updateData data.UpdateUser) error {
	updates := make(map[string]any)
	updates["name"] = updateData.Name

	if len(updates) == 0 {
		return nil
	}

	return update.repository.Update(user, updates)
}
