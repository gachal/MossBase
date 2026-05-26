package dto

import "time"

type CreatePageRequest struct {
	Title    string  `json:"title" binding:"required,max=200"`
	Content  string  `json:"content"`
	ParentID *uint64 `json:"parent_id"`
	Position *int    `json:"position"`
}

type UpdatePageRequest struct {
	Title       *string `json:"title" binding:"max=200"`
	Content     *string `json:"content"`
	ContentHTML *string `json:"content_html"`
}

type MovePageRequest struct {
	ParentID *uint64 `json:"parent_id"`
	Position *int    `json:"position"`
}

type PageResponse struct {
	ID          uint64     `json:"id"`
	SpaceID     uint64     `json:"space_id"`
	ParentID    *uint64    `json:"parent_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content,omitempty"`
	ContentHTML string     `json:"content_html,omitempty"`
	Position    int        `json:"position"`
	Status      string     `json:"status"`
	Version     int        `json:"version"`
	CreatedBy   uint64     `json:"created_by"`
	UpdatedBy   uint64     `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PageTreeResponse struct {
	ID          uint64              `json:"id"`
	SpaceID     uint64              `json:"space_id"`
	ParentID    *uint64             `json:"parent_id"`
	Title       string              `json:"title"`
	Slug        string              `json:"slug"`
	Position    int                 `json:"position"`
	Status      string              `json:"status"`
	Version     int                 `json:"version"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Children    []*PageTreeResponse `json:"children"`
}

type PageVersionResponse struct {
	ID            uint64    `json:"id"`
	PageID        uint64    `json:"page_id"`
	VersionNumber int       `json:"version_number"`
	Title         string    `json:"title"`
	Content       string    `json:"content,omitempty"`
	ContentHTML   string    `json:"content_html,omitempty"`
	EditedBy      uint64    `json:"edited_by"`
	CreatedAt     time.Time `json:"created_at"`
}

type VersionDiffRequest struct {
	From int `form:"from" binding:"required"`
	To   int `form:"to" binding:"required"`
}

type VersionDiffResponse struct {
	FromVersion int    `json:"from_version"`
	ToVersion   int    `json:"to_version"`
	Diff        string `json:"diff"`
}

type SearchResultItem struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Snippet   string    `json:"snippet"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SearchResultResponse struct {
	Items    []SearchResultItem `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

type SemanticSearchResponse struct {
	Results []SemanticSearchItem `json:"results"`
	Total   int                  `json:"total"`
	Query   string               `json:"query"`
}

type SemanticSearchItem struct {
	PageID  uint64  `json:"page_id"`
	Title   string  `json:"title"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
}
