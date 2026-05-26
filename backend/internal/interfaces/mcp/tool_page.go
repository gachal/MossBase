package mcp

import (
	"context"
	"fmt"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxPageSize = 100

// PageToolHandler provides MCP tool handlers for page operations.
type PageToolHandler struct {
	pageSvc service.PageService
	auth    *MCPAuth
	authz   *spaceAuthorizer
}

// NewPageToolHandler creates a new PageToolHandler with the given dependencies.
func NewPageToolHandler(pageSvc service.PageService, auth *MCPAuth, authz *spaceAuthorizer) *PageToolHandler {
	return &PageToolHandler{pageSvc: pageSvc, auth: auth, authz: authz}
}

// CreatePage creates a new page in the specified space.
func (h *PageToolHandler) CreatePage(ctx context.Context, req *mcpsdk.CallToolRequest, input CreatePageInput) (*mcpsdk.CallToolResult, PageOutput, error) {
	if err := h.authz.checkWrite(ctx, input.SpaceID); err != nil {
		return nil, PageOutput{}, err
	}

	resp, err := h.pageSvc.Create(ctx, input.SpaceID, h.auth.UserID(), dto.CreatePageRequest{
		Title:    input.Title,
		Content:  input.Content,
		ParentID: input.ParentID,
	})
	if err != nil {
		return nil, PageOutput{}, fmt.Errorf("failed to create page: %w", err)
	}

	return nil, toPageOutput(resp), nil
}

// GetPage retrieves a single page by ID.
func (h *PageToolHandler) GetPage(ctx context.Context, req *mcpsdk.CallToolRequest, input GetPageInput) (*mcpsdk.CallToolResult, PageOutput, error) {
	resp, err := h.pageSvc.GetByID(ctx, input.PageID)
	if err != nil {
		return nil, PageOutput{}, fmt.Errorf("failed to get page %d: %w", input.PageID, err)
	}

	if err := h.authz.checkRead(ctx, resp.SpaceID); err != nil {
		return nil, PageOutput{}, err
	}

	return nil, toPageOutput(resp), nil
}

// UpdatePage updates the title and/or content of an existing page.
func (h *PageToolHandler) UpdatePage(ctx context.Context, req *mcpsdk.CallToolRequest, input UpdatePageInput) (*mcpsdk.CallToolResult, PageOutput, error) {
	existing, err := h.pageSvc.GetByID(ctx, input.PageID)
	if err != nil {
		return nil, PageOutput{}, fmt.Errorf("failed to get page %d: %w", input.PageID, err)
	}

	if err := h.authz.checkWrite(ctx, existing.SpaceID); err != nil {
		return nil, PageOutput{}, err
	}

	resp, err := h.pageSvc.Update(ctx, input.PageID, h.auth.UserID(), dto.UpdatePageRequest{
		Title:   input.Title,
		Content: input.Content,
	})
	if err != nil {
		return nil, PageOutput{}, fmt.Errorf("failed to update page %d: %w", input.PageID, err)
	}

	return nil, toPageOutput(resp), nil
}

// DeletePage deletes a page by ID.
func (h *PageToolHandler) DeletePage(ctx context.Context, req *mcpsdk.CallToolRequest, input DeletePageInput) (*mcpsdk.CallToolResult, DeleteOutput, error) {
	existing, err := h.pageSvc.GetByID(ctx, input.PageID)
	if err != nil {
		return nil, DeleteOutput{}, fmt.Errorf("failed to get page %d: %w", input.PageID, err)
	}

	if err := h.authz.checkWrite(ctx, existing.SpaceID); err != nil {
		return nil, DeleteOutput{}, err
	}

	if err := h.pageSvc.Delete(ctx, input.PageID); err != nil {
		return nil, DeleteOutput{}, fmt.Errorf("failed to delete page %d: %w", input.PageID, err)
	}

	return nil, DeleteOutput{Success: true, Message: "page deleted"}, nil
}

// MovePage moves a page to a new parent and/or position.
func (h *PageToolHandler) MovePage(ctx context.Context, req *mcpsdk.CallToolRequest, input MovePageInput) (*mcpsdk.CallToolResult, PageOutput, error) {
	existing, err := h.pageSvc.GetByID(ctx, input.PageID)
	if err != nil {
		return nil, PageOutput{}, fmt.Errorf("failed to get page %d: %w", input.PageID, err)
	}

	if err := h.authz.checkWrite(ctx, existing.SpaceID); err != nil {
		return nil, PageOutput{}, err
	}

	resp, err := h.pageSvc.MovePage(ctx, input.PageID, dto.MovePageRequest{
		ParentID: input.ParentID,
		Position: input.Position,
	})
	if err != nil {
		return nil, PageOutput{}, fmt.Errorf("failed to move page %d: %w", input.PageID, err)
	}

	return nil, toPageOutput(resp), nil
}

// GetPageTree returns the full page tree for a space.
func (h *PageToolHandler) GetPageTree(ctx context.Context, req *mcpsdk.CallToolRequest, input GetPageTreeInput) (*mcpsdk.CallToolResult, []*PageTreeOutput, error) {
	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, nil, err
	}

	nodes, err := h.pageSvc.GetTreeBySpace(ctx, input.SpaceID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get page tree for space %d: %w", input.SpaceID, err)
	}

	return nil, toPageTreeOutputs(nodes), nil
}
