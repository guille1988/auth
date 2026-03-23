package model

import (
	"api/internal/infrastructure/config"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UUID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"uuid"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Email     string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// ConnectionName specifies the database connection name for the User model
func (User) ConnectionName() config.ConnectionName {
	return config.Default
}
