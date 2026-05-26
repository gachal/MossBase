package model

import (
	"time"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"gorm.io/gorm"
)

type PageModel struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SpaceID     uint64         `gorm:"index;not null" json:"space_id"`
	ParentID    *uint64        `gorm:"column:parent_id;index" json:"parent_id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Slug        string         `gorm:"size:255" json:"slug"`
	Content     string         `gorm:"type:longtext" json:"content"`
	ContentHTML string         `gorm:"column:content_html;type:longtext" json:"content_html"`
	Position    int            `gorm:"default:0" json:"position"`
	Status      string         `gorm:"size:20;default:draft;not null" json:"status"`
	Version     int            `gorm:"default:1" json:"version"`
	CreatedBy   uint64         `gorm:"column:created_by;not null" json:"created_by"`
	UpdatedBy   uint64         `gorm:"column:updated_by;not null" json:"updated_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PageModel) TableName() string { return "pages" }

func (m PageModel) ToEntity() *entity.Page {
	return &entity.Page{
		ID:          m.ID,
		SpaceID:     m.SpaceID,
		ParentID:    m.ParentID,
		Title:       m.Title,
		Slug:        m.Slug,
		Content:     m.Content,
		ContentHTML: m.ContentHTML,
		Position:    m.Position,
		Status:      entity.PageStatus(m.Status),
		Version:     m.Version,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func FromPageEntity(e *entity.Page) *PageModel {
	return &PageModel{
		ID:          e.ID,
		SpaceID:     e.SpaceID,
		ParentID:    e.ParentID,
		Title:       e.Title,
		Slug:        e.Slug,
		Content:     e.Content,
		ContentHTML: e.ContentHTML,
		Position:    e.Position,
		Status:      string(e.Status),
		Version:     e.Version,
		CreatedBy:   e.CreatedBy,
		UpdatedBy:   e.UpdatedBy,
	}
}
