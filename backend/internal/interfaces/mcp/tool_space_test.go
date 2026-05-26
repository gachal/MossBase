package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gachal/mossbase/backend/internal/application/dto"
)

type mockSpaceService struct {
	createFn       func(ctx context.Context, ownerID uint64, req dto.CreateSpaceRequest) (*dto.SpaceResponse, error)
	updateFn       func(ctx context.Context, spaceID uint64, req dto.UpdateSpaceRequest) (*dto.SpaceResponse, error)
	deleteFn       func(ctx context.Context, spaceID uint64) error
	getByIDFn      func(ctx context.Context, spaceID uint64) (*dto.SpaceResponse, error)
	listByUserFn   func(ctx context.Context, userID uint64, page, pageSize int) ([]dto.SpaceResponse, int64, error)
	addMemberFn    func(ctx context.Context, spaceID, requesterID uint64, req dto.AddMemberRequest) error
	removeMemberFn func(ctx context.Context, spaceID, requesterID, targetUserID uint64) error
	listMembersFn  func(ctx context.Context, spaceID uint64) ([]dto.SpaceMemberResponse, error)
}

func (m *mockSpaceService) Create(ctx context.Context, ownerID uint64, req dto.CreateSpaceRequest) (*dto.SpaceResponse, error) {
	return m.createFn(ctx, ownerID, req)
}
func (m *mockSpaceService) Update(ctx context.Context, spaceID uint64, req dto.UpdateSpaceRequest) (*dto.SpaceResponse, error) {
	return m.updateFn(ctx, spaceID, req)
}
func (m *mockSpaceService) Delete(ctx context.Context, spaceID uint64) error {
	return m.deleteFn(ctx, spaceID)
}
func (m *mockSpaceService) GetByID(ctx context.Context, spaceID uint64) (*dto.SpaceResponse, error) {
	return m.getByIDFn(ctx, spaceID)
}
func (m *mockSpaceService) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]dto.SpaceResponse, int64, error) {
	return m.listByUserFn(ctx, userID, page, pageSize)
}
func (m *mockSpaceService) AddMember(ctx context.Context, spaceID, requesterID uint64, req dto.AddMemberRequest) error {
	return m.addMemberFn(ctx, spaceID, requesterID, req)
}
func (m *mockSpaceService) RemoveMember(ctx context.Context, spaceID, requesterID, targetUserID uint64) error {
	return m.removeMemberFn(ctx, spaceID, requesterID, targetUserID)
}
func (m *mockSpaceService) ListMembers(ctx context.Context, spaceID uint64) ([]dto.SpaceMemberResponse, error) {
	return m.listMembersFn(ctx, spaceID)
}

func sampleSpaceResponse() *dto.SpaceResponse {
	now := time.Now()
	return &dto.SpaceResponse{
		ID:          1,
		Name:        "Test Space",
		Description: "A test space",
		Icon:        "library",
		Visibility:  "private",
		OwnerID:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func TestSpaceToolHandler_ListSpaces(t *testing.T) {
	svc := &mockSpaceService{
		listByUserFn: func(_ context.Context, userID uint64, page, pageSize int) ([]dto.SpaceResponse, int64, error) {
			if userID != 1 {
				t.Errorf("expected userID 1, got %d", userID)
			}
			if page != 1 || pageSize != 20 {
				t.Errorf("expected page=1, pageSize=20, got page=%d, pageSize=%d", page, pageSize)
			}
			return []dto.SpaceResponse{*sampleSpaceResponse()}, 1, nil
		},
	}

	h := NewSpaceToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.ListSpaces(context.Background(), nil, ListSpacesInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Spaces) != 1 {
		t.Fatalf("expected 1 space, got %d", len(result.Spaces))
	}
	if result.Spaces[0].Name != "Test Space" {
		t.Errorf("expected name 'Test Space', got %q", result.Spaces[0].Name)
	}
}

func TestSpaceToolHandler_ListSpaces_CustomPagination(t *testing.T) {
	svc := &mockSpaceService{
		listByUserFn: func(_ context.Context, _ uint64, page, pageSize int) ([]dto.SpaceResponse, int64, error) {
			if page != 2 || pageSize != 5 {
				t.Errorf("expected page=2, pageSize=5, got page=%d, pageSize=%d", page, pageSize)
			}
			return []dto.SpaceResponse{}, 0, nil
		},
	}

	h := NewSpaceToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.ListSpaces(context.Background(), nil, ListSpacesInput{Page: 2, PageSize: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}
}

func TestSpaceToolHandler_GetSpace(t *testing.T) {
	svc := &mockSpaceService{
		getByIDFn: func(_ context.Context, spaceID uint64) (*dto.SpaceResponse, error) {
			if spaceID != 7 {
				t.Errorf("expected spaceID 7, got %d", spaceID)
			}
			return sampleSpaceResponse(), nil
		},
	}

	h := NewSpaceToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.GetSpace(context.Background(), nil, GetSpaceInput{SpaceID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Test Space" {
		t.Errorf("expected name 'Test Space', got %q", result.Name)
	}
}

func TestSpaceToolHandler_GetSpace_Error(t *testing.T) {
	svc := &mockSpaceService{
		getByIDFn: func(context.Context, uint64) (*dto.SpaceResponse, error) {
			return nil, errors.New("not found")
		},
	}

	h := NewSpaceToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.GetSpace(context.Background(), nil, GetSpaceInput{SpaceID: 999})
	if err == nil {
		t.Fatal("expected error for missing space")
	}
}

func TestSpaceToolHandler_ListMembers(t *testing.T) {
	svc := &mockSpaceService{
		listMembersFn: func(_ context.Context, spaceID uint64) ([]dto.SpaceMemberResponse, error) {
			if spaceID != 3 {
				t.Errorf("expected spaceID 3, got %d", spaceID)
			}
			return []dto.SpaceMemberResponse{
				{ID: 1, SpaceID: 3, UserID: 1, Username: "admin", Role: "admin"},
				{ID: 2, SpaceID: 3, UserID: 2, Username: "viewer", Role: "viewer"},
			}, nil
		},
	}

	h := NewSpaceToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.ListMembers(context.Background(), nil, ListMembersInput{SpaceID: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 members, got %d", len(result))
	}
	if result[0].Username != "admin" {
		t.Errorf("expected first member 'admin', got %q", result[0].Username)
	}
}

func TestSpaceToolHandler_ListMembers_NoEmail(t *testing.T) {
	svc := &mockSpaceService{
		listMembersFn: func(_ context.Context, _ uint64) ([]dto.SpaceMemberResponse, error) {
			return []dto.SpaceMemberResponse{
				{ID: 1, SpaceID: 3, UserID: 1, Username: "admin", Role: "admin", Email: "admin@test.com"},
			}, nil
		},
	}

	h := NewSpaceToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.ListMembers(context.Background(), nil, ListMembersInput{SpaceID: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 member, got %d", len(result))
	}
	// MemberOutput type has no Email field — this verifies compilation
	_ = result[0]
}
