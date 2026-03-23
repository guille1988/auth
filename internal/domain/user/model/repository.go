package model

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]User, error)
	FindByID(id uint) (*User, error)
	FindByUUID(uuid string) (*User, error)
	FindByEmail(email string) (*User, error)
	ExistByEmail(email string) (bool, error)
	Create(user *User) error
	CreateMany(users []User) error
	Update(user *User, data map[string]any) error
	Delete(user *User) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (repo *repository) FindAll() ([]User, error) {
	var users []User

	result := repo.db.Order("id desc").Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

func (repo *repository) ExistByEmail(email string) (bool, error) {
	var exists bool

	err := repo.db.Model(&User{}).
		Select("count(*) > 0").
		Where("email = ?", email).
		Find(&exists).
		Error

	return exists, err
}

func (repo *repository) FindByID(id uint) (*User, error) {
	var user User

	result := repo.db.First(&user, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (repo *repository) FindByUUID(uuid string) (*User, error) {
	var user User

	result := repo.db.Where("uuid = ?", uuid).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (repo *repository) FindByEmail(email string) (*User, error) {
	var user User

	result := repo.db.Where("email = ?", email).First(&user)

	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (repo *repository) Create(user *User) error {
	return repo.db.Create(user).Error
}

func (repo *repository) CreateMany(users []User) error {
	return repo.db.Create(&users).Error
}

func (repo *repository) Update(user *User, data map[string]any) error {
	return repo.db.Model(user).Updates(data).Error
}

func (repo *repository) Delete(user *User) error {
	return repo.db.Delete(user).Error
}
