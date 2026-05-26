package mcp

import (
	"context"
	"fmt"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	maxLimit   = 50
	maxQueryLen = 500
)

// SearchToolHandler provides MCP tool handlers for search operations.
type SearchToolHandler struct {
	pageSvc service.PageService
	auth    *MCPAuth
	authz   *spaceAuthorizer
}

// NewSearchToolHandler creates a new SearchToolHandler with the given dependencies.
func NewSearchToolHandler(pageSvc service.PageService, auth *MCPAuth, authz *spaceAuthorizer) *SearchToolHandler {
	return &SearchToolHandler{pageSvc: pageSvc, auth: auth, authz: authz}
}

// Search performs a full-text search for pages within a space.
func (h *SearchToolHandler) Search(ctx context.Context, req *mcpsdk.CallToolRequest, input SearchInput) (*mcpsdk.CallToolResult, *dto.SearchResultResponse, error) {
	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, nil, err
	}
	if len(input.Query) == 0 || len(input.Query) > maxQueryLen {
		return nil, nil, fmt.Errorf("query length must be between 1 and %d characters", maxQueryLen)
	}

	page := input.Page
	if page == 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	result, err := h.pageSvc.Search(ctx, input.SpaceID, input.Query, page, pageSize)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to search in space %d: %w", input.SpaceID, err)
	}

	return nil, result, nil
}

// SemanticSearch performs a semantic (vector-based) search for pages within a space.
func (h *SearchToolHandler) SemanticSearch(ctx context.Context, req *mcpsdk.CallToolRequest, input SemanticSearchInput) (*mcpsdk.CallToolResult, *dto.SemanticSearchResponse, error) {
	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, nil, err
	}
	if len(input.Query) == 0 || len(input.Query) > maxQueryLen {
		return nil, nil, fmt.Errorf("query length must be between 1 and %d characters", maxQueryLen)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 10
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	result, err := h.pageSvc.SemanticSearch(ctx, input.SpaceID, input.Query, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to semantic search in space %d: %w", input.SpaceID, err)
	}

	return nil, result, nil
}
