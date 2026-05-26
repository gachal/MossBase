package model

import (
	"time"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"gorm.io/gorm"
)

type UserModel struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Username     string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash string         `gorm:"column:password_hash;size:255;not null" json:"-"`
	Avatar       string         `gorm:"size:500" json:"avatar"`
	Role         string         `gorm:"size:20;default:user;not null" json:"role"`
	Status       string         `gorm:"size:20;default:active;not null" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserModel) TableName() string { return "users" }

func (m UserModel) ToEntity() *entity.User {
	return &entity.User{
		ID:           m.ID,
		Email:        m.Email,
		Username:     m.Username,
		PasswordHash: m.PasswordHash,
		Avatar:       m.Avatar,
		Role:         entity.UserRole(m.Role),
		Status:       entity.UserStatus(m.Status),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func FromUserEntity(e *entity.User) *UserModel {
	return &UserModel{
		ID:           e.ID,
		Email:        e.Email,
		Username:     e.Username,
		PasswordHash: e.PasswordHash,
		Avatar:       e.Avatar,
		Role:         string(e.Role),
		Status:       string(e.Status),
	}
}
