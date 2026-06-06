package mcp

import (
	"time"

	"github.com/gachal/mossbase/backend/internal/application/dto"
)

// --- Page Tool Inputs ---

type CreatePageInput struct {
	SpaceID  uint64  `json:"space_id"  jsonschema:"目标空间 ID"`
	Title    string  `json:"title"     jsonschema:"页面标题"`
	Content  string  `json:"content"   jsonschema:"页面 Markdown 内容"`
	ParentID *uint64 `json:"parent_id" jsonschema:"父页面 ID，为空则作为根页面"`
}

type GetPageInput struct {
	PageID uint64 `json:"page_id" jsonschema:"页面 ID"`
}

type UpdatePageInput struct {
	PageID  uint64  `json:"page_id"  jsonschema:"页面 ID"`
	Title   *string `json:"title"    jsonschema:"新标题"`
	Content *string `json:"content"  jsonschema:"新 Markdown 内容"`
}

type DeletePageInput struct {
	PageID uint64 `json:"page_id" jsonschema:"页面 ID"`
}

type MovePageInput struct {
	PageID   uint64  `json:"page_id"   jsonschema:"页面 ID"`
	ParentID *uint64 `json:"parent_id" jsonschema:"新父页面 ID，0 表示根级"`
	Position *int    `json:"position"  jsonschema:"在同级中的位置"`
}

type GetPageTreeInput struct {
	SpaceID uint64 `json:"space_id" jsonschema:"空间 ID"`
}

// --- Space Tool Inputs ---

type ListSpacesInput struct {
	Page     int `json:"page"      jsonschema:"页码，默认 1"`
	PageSize int `json:"page_size" jsonschema:"每页数量，默认 20"`
}

type GetSpaceInput struct {
	SpaceID uint64 `json:"space_id" jsonschema:"空间 ID"`
}

type ListMembersInput struct {
	SpaceID uint64 `json:"space_id" jsonschema:"空间 ID"`
}

// --- Search Tool Inputs ---

type SearchInput struct {
	SpaceID  uint64 `json:"space_id"  jsonschema:"空间 ID"`
	Query    string `json:"query"     jsonschema:"搜索关键词"`
	Page     int    `json:"page"      jsonschema:"页码，默认 1"`
	PageSize int    `json:"page_size" jsonschema:"每页数量，默认 20"`
}

type SemanticSearchInput struct {
	SpaceID uint64 `json:"space_id" jsonschema:"空间 ID"`
	Query   string `json:"query"    jsonschema:"搜索查询"`
	Limit   int    `json:"limit"    jsonschema:"最大结果数，默认 10"`
}

// --- Output Types ---

type PageOutput struct {
	ID        uint64    `json:"id"`
	SpaceID   uint64    `json:"space_id"`
	ParentID  *uint64   `json:"parent_id,omitempty"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Content   string    `json:"content,omitempty"`
	Status    string    `json:"status"`
	Version   int       `json:"version"`
	CreatedBy uint64    `json:"created_by"`
	UpdatedBy uint64    `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PageTreeOutput struct {
	ID       uint64 `json:"id"`
	SpaceID  uint64 `json:"space_id"`
	ParentID uint64 `json:"parent_id"`
	Title    string `json:"title"`
	Slug     string `json:"slug"`
	Status   string `json:"status"`
}

type PageTreeResult struct {
	Pages []PageTreeOutput `json:"pages"`
}

type SpaceOutput struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Visibility  string    `json:"visibility"`
	OwnerID     uint64    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DeleteOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// --- Converters ---

func toPageOutput(p *dto.PageResponse) PageOutput {
	return PageOutput{
		ID:        p.ID,
		SpaceID:   p.SpaceID,
		ParentID:  p.ParentID,
		Title:     p.Title,
		Slug:      p.Slug,
		Content:   p.Content,
		Status:    p.Status,
		Version:   p.Version,
		CreatedBy: p.CreatedBy,
		UpdatedBy: p.UpdatedBy,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toPageTreeOutputs(nodes []*dto.PageTreeResponse) []PageTreeOutput {
	var result []PageTreeOutput
	var flatten func([]*dto.PageTreeResponse)
	flatten = func(items []*dto.PageTreeResponse) {
		for _, n := range items {
			result = append(result, PageTreeOutput{
				ID:       n.ID,
				Title:    n.Title,
				Slug:     n.Slug,
				Status:   n.Status,
			})
			if len(n.Children) > 0 {
				flatten(n.Children)
			}
		}
	}
	flatten(nodes)
	return result
}
