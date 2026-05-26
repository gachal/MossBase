package model

import (
	"time"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"gorm.io/gorm"
)

type SpaceModel struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Description string         `gorm:"size:500" json:"description"`
	Icon        string         `gorm:"size:500" json:"icon"`
	Visibility  string         `gorm:"size:20;default:private;not null" json:"visibility"`
	OwnerID     uint64         `gorm:"column:owner_id;index;not null" json:"owner_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SpaceModel) TableName() string { return "spaces" }

func (m SpaceModel) ToEntity() *entity.Space {
	return &entity.Space{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Icon:        m.Icon,
		Visibility:  entity.SpaceVisibility(m.Visibility),
		OwnerID:     m.OwnerID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func FromSpaceEntity(e *entity.Space) *SpaceModel {
	return &SpaceModel{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Icon:        e.Icon,
		Visibility:  string(e.Visibility),
		OwnerID:     e.OwnerID,
	}
}
