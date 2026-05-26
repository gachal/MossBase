package entity

import "time"

type PageStatus string

const (
	PageStatusDraft     PageStatus = "draft"
	PageStatusPublished PageStatus = "published"
)

type Page struct {
	ID          uint64
	SpaceID     uint64
	ParentID    *uint64
	Title       string
	Slug        string
	Content     string
	ContentHTML string
	Position    int
	Status      PageStatus
	Version     int
	CreatedBy   uint64
	UpdatedBy   uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p Page) IsRoot() bool {
	return p.ParentID == nil
}

type PageTreeNode struct {
	Page
	Children []*PageTreeNode
}
