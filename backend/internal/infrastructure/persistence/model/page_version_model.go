package model

import (
	"time"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type PageVersionModel struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PageID        uint64    `gorm:"index;column:page_id;not null" json:"page_id"`
	VersionNumber int       `gorm:"column:version_number;not null" json:"version_number"`
	Title         string    `gorm:"size:255" json:"title"`
	Content       string    `gorm:"type:longtext" json:"content"`
	ContentHTML   string    `gorm:"column:content_html;type:longtext" json:"content_html"`
	EditedBy      uint64    `gorm:"column:edited_by;not null" json:"edited_by"`
	CreatedAt     time.Time `json:"created_at"`
}

func (PageVersionModel) TableName() string { return "page_versions" }

func (m PageVersionModel) ToEntity() *entity.PageVersion {
	return &entity.PageVersion{
		ID:            m.ID,
		PageID:        m.PageID,
		VersionNumber: m.VersionNumber,
		Title:         m.Title,
		Content:       m.Content,
		ContentHTML:   m.ContentHTML,
		EditedBy:      m.EditedBy,
		CreatedAt:     m.CreatedAt,
	}
}

func FromPageVersionEntity(e *entity.PageVersion) *PageVersionModel {
	return &PageVersionModel{
		ID:            e.ID,
		PageID:        e.PageID,
		VersionNumber: e.VersionNumber,
		Title:         e.Title,
		Content:       e.Content,
		ContentHTML:   e.ContentHTML,
		EditedBy:      e.EditedBy,
	}
}
