package models

import (
	"time"

	"github.com/EmanuelErnesto/uninaquiz-backend/internal/domain/entities"
)

type UserModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	Username  string    `gorm:"column:username;uniqueIndex;not null"`
	Password  string    `gorm:"column:password;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserModel) TableName() string { return "tb_users" }

func (m *UserModel) ToDomain() *entities.User {
	return &entities.User{
		ID:        m.ID,
		Username:  m.Username,
		Password:  m.Password,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func UserToModel(u entities.User) *UserModel {
	return &UserModel{
		ID:        u.ID,
		Username:  u.Username,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
