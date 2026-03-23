package model

import (
	"errors"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Factory struct {
	repository     Repository
	hashedPassword string
}

func NewFactory(db *gorm.DB) *Factory {
	repo := NewRepository(db)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	return &Factory{
		repository:     repo,
		hashedPassword: string(hashedPassword),
	}
}

func newUser(hashedPassword string) User {
	return User{
		UUID:     uuid.New(),
		Name:     gofakeit.Name(),
		Email:    gofakeit.Email(),
		Password: hashedPassword,
	}
}

func newUsers(quantity int, hashedPassword string) []User {
	users := make([]User, quantity)

	for i := range quantity {
		users[i] = newUser(hashedPassword)
	}

	return users
}

func (factory *Factory) Create(quantity int) ([]User, error) {
	if quantity <= 0 {
		return nil, errors.New("the number of users must be greater than 0")
	}

	if quantity == 1 {
		user := newUser(factory.hashedPassword)
		err := factory.repository.Create(&user)

		if err != nil {
			return nil, err
		}

		return []User{user}, nil
	}

	users := newUsers(quantity, factory.hashedPassword)
	err := factory.repository.CreateMany(users)

	if err != nil {
		return nil, err
	}

	return users, nil
}
