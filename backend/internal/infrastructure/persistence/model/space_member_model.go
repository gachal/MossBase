package model

import (
	"time"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type SpaceMemberModel struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SpaceID   uint64    `gorm:"uniqueIndex:idx_space_user;not null" json:"space_id"`
	UserID    uint64    `gorm:"uniqueIndex:idx_space_user;not null" json:"user_id"`
	Role      string    `gorm:"size:20;default:member;not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func (SpaceMemberModel) TableName() string { return "space_members" }

func (m SpaceMemberModel) ToEntity() *entity.SpaceMember {
	return &entity.SpaceMember{
		ID:        m.ID,
		SpaceID:   m.SpaceID,
		UserID:    m.UserID,
		Role:      entity.MemberRole(m.Role),
		CreatedAt: m.CreatedAt,
	}
}

func FromSpaceMemberEntity(e *entity.SpaceMember) *SpaceMemberModel {
	return &SpaceMemberModel{
		ID:      e.ID,
		SpaceID: e.SpaceID,
		UserID:  e.UserID,
		Role:    string(e.Role),
	}
}
