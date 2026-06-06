package mcp

import (
	"context"
	"fmt"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListSpacesOutput is the output type for ListSpaces.
type ListSpacesOutput struct {
	Spaces []SpaceOutput `json:"spaces"`
	Total  int64         `json:"total"`
}

// MemberOutput represents a space member without sensitive fields like email.
type MemberOutput struct {
	ID       uint64 `json:"id"`
	SpaceID  uint64 `json:"space_id"`
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// SpaceToolHandler provides MCP tool handlers for space operations.
type SpaceToolHandler struct {
	spaceSvc service.SpaceService
	auth     *MCPAuth
	authz    *spaceAuthorizer
}

// NewSpaceToolHandler creates a new SpaceToolHandler with the given dependencies.
func NewSpaceToolHandler(spaceSvc service.SpaceService, auth *MCPAuth, authz *spaceAuthorizer) *SpaceToolHandler {
	return &SpaceToolHandler{spaceSvc: spaceSvc, auth: auth, authz: authz}
}

// ListSpaces lists spaces accessible to the current user with pagination.
func (h *SpaceToolHandler) ListSpaces(ctx context.Context, req *mcpsdk.CallToolRequest, input ListSpacesInput) (*mcpsdk.CallToolResult, ListSpacesOutput, error) {
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

	spaces, total, err := h.spaceSvc.ListByUser(ctx, h.auth.UserID(), page, pageSize)
	if err != nil {
		return nil, ListSpacesOutput{}, fmt.Errorf("failed to list spaces: %w", err)
	}

	outputs := make([]SpaceOutput, len(spaces))
	for i, sp := range spaces {
		outputs[i] = toSpaceOutput(&sp)
	}

	return nil, ListSpacesOutput{
		Spaces: outputs,
		Total:  total,
	}, nil
}

// GetSpace retrieves a single space by ID.
func (h *SpaceToolHandler) GetSpace(ctx context.Context, req *mcpsdk.CallToolRequest, input GetSpaceInput) (*mcpsdk.CallToolResult, SpaceOutput, error) {
	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, SpaceOutput{}, err
	}

	resp, err := h.spaceSvc.GetByID(ctx, input.SpaceID)
	if err != nil {
		return nil, SpaceOutput{}, fmt.Errorf("failed to get space %d: %w", input.SpaceID, err)
	}

	return nil, toSpaceOutput(resp), nil
}

// MemberListResult wraps member list for SDK output schema compatibility.
type MemberListResult struct {
	Members []MemberOutput `json:"members"`
}

// ListMembers lists all members of a space (without email).
func (h *SpaceToolHandler) ListMembers(ctx context.Context, req *mcpsdk.CallToolRequest, input ListMembersInput) (*mcpsdk.CallToolResult, MemberListResult, error) {
	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, MemberListResult{}, err
	}

	members, err := h.spaceSvc.ListMembers(ctx, input.SpaceID)
	if err != nil {
		return nil, MemberListResult{}, fmt.Errorf("failed to list members for space %d: %w", input.SpaceID, err)
	}

	outputs := make([]MemberOutput, len(members))
	for i, m := range members {
		outputs[i] = MemberOutput{
			ID:       m.ID,
			SpaceID:  m.SpaceID,
			UserID:   m.UserID,
			Username: m.Username,
			Role:     m.Role,
		}
	}

	return nil, MemberListResult{Members: outputs}, nil
}

// toSpaceOutput converts a dto.SpaceResponse to SpaceOutput.
func toSpaceOutput(sp *dto.SpaceResponse) SpaceOutput {
	return SpaceOutput{
		ID:          sp.ID,
		Name:        sp.Name,
		Description: sp.Description,
		Icon:        sp.Icon,
		Visibility:  sp.Visibility,
		OwnerID:     sp.OwnerID,
		CreatedAt:   sp.CreatedAt,
		UpdatedAt:   sp.UpdatedAt,
	}
}
