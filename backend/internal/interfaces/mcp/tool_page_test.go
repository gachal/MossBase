package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type mockSpaceMemberRepo struct {
	findBySpaceAndUserFn func(ctx context.Context, spaceID, userID uint64) (*entity.SpaceMember, error)
}

func (m *mockSpaceMemberRepo) Create(ctx context.Context, member *entity.SpaceMember) error {
	panic("unimplemented")
}
func (m *mockSpaceMemberRepo) FindBySpaceAndUser(ctx context.Context, spaceID, userID uint64) (*entity.SpaceMember, error) {
	return m.findBySpaceAndUserFn(ctx, spaceID, userID)
}
func (m *mockSpaceMemberRepo) FindBySpaceID(ctx context.Context, spaceID uint64) ([]entity.SpaceMember, error) {
	panic("unimplemented")
}
func (m *mockSpaceMemberRepo) FindByUserID(ctx context.Context, userID uint64) ([]entity.SpaceMember, error) {
	panic("unimplemented")
}
func (m *mockSpaceMemberRepo) Delete(ctx context.Context, spaceID, userID uint64) error {
	panic("unimplemented")
}
func (m *mockSpaceMemberRepo) CountAdminsBySpaceID(ctx context.Context, spaceID uint64) (int64, error) {
	panic("unimplemented")
}

func allowAllAuthz(t *testing.T) *spaceAuthorizer {
	t.Helper()
	memberRepo := &mockSpaceMemberRepo{
		findBySpaceAndUserFn: func(_ context.Context, _, _ uint64) (*entity.SpaceMember, error) {
			return &entity.SpaceMember{ID: 1, SpaceID: 1, UserID: 1, Role: entity.MemberRoleAdmin}, nil
		},
	}
	return &spaceAuthorizer{memberRepo: memberRepo, auth: NewMCPAuth(nil, 1)}
}

func denyAuthz(t *testing.T) *spaceAuthorizer {
	t.Helper()
	memberRepo := &mockSpaceMemberRepo{
		findBySpaceAndUserFn: func(_ context.Context, _, _ uint64) (*entity.SpaceMember, error) {
			return nil, errors.New("not a member")
		},
	}
	return &spaceAuthorizer{memberRepo: memberRepo, auth: NewMCPAuth(nil, 1)}
}

type mockPageService struct {
	createFn         func(ctx context.Context, spaceID, userID uint64, req dto.CreatePageRequest) (*dto.PageResponse, error)
	getByIDFn        func(ctx context.Context, pageID uint64) (*dto.PageResponse, error)
	updateFn         func(ctx context.Context, pageID, userID uint64, req dto.UpdatePageRequest) (*dto.PageResponse, error)
	deleteFn         func(ctx context.Context, pageID uint64) error
	getTreeBySpaceFn func(ctx context.Context, spaceID uint64) ([]*dto.PageTreeResponse, error)
	movePageFn       func(ctx context.Context, pageID uint64, req dto.MovePageRequest) (*dto.PageResponse, error)
	searchFn         func(ctx context.Context, spaceID uint64, query string, page, pageSize int) (*dto.SearchResultResponse, error)
	semanticSearchFn func(ctx context.Context, spaceID uint64, query string, limit int) (*dto.SemanticSearchResponse, error)
}

func (m *mockPageService) Create(ctx context.Context, spaceID, userID uint64, req dto.CreatePageRequest) (*dto.PageResponse, error) {
	return m.createFn(ctx, spaceID, userID, req)
}
func (m *mockPageService) GetByID(ctx context.Context, pageID uint64) (*dto.PageResponse, error) {
	return m.getByIDFn(ctx, pageID)
}
func (m *mockPageService) Update(ctx context.Context, pageID, userID uint64, req dto.UpdatePageRequest) (*dto.PageResponse, error) {
	return m.updateFn(ctx, pageID, userID, req)
}
func (m *mockPageService) Delete(ctx context.Context, pageID uint64) error {
	return m.deleteFn(ctx, pageID)
}
func (m *mockPageService) GetTreeBySpace(ctx context.Context, spaceID uint64) ([]*dto.PageTreeResponse, error) {
	return m.getTreeBySpaceFn(ctx, spaceID)
}
func (m *mockPageService) MovePage(ctx context.Context, pageID uint64, req dto.MovePageRequest) (*dto.PageResponse, error) {
	return m.movePageFn(ctx, pageID, req)
}
func (m *mockPageService) Search(ctx context.Context, spaceID uint64, query string, page, pageSize int) (*dto.SearchResultResponse, error) {
	return m.searchFn(ctx, spaceID, query, page, pageSize)
}
func (m *mockPageService) SemanticSearch(ctx context.Context, spaceID uint64, query string, limit int) (*dto.SemanticSearchResponse, error) {
	return m.semanticSearchFn(ctx, spaceID, query, limit)
}

func samplePageResponse() *dto.PageResponse {
	now := time.Now()
	return &dto.PageResponse{
		ID:        1,
		SpaceID:   10,
		Title:     "Test Page",
		Slug:      "test-page",
		Content:   "# Hello",
		Status:    "published",
		Version:   2,
		CreatedBy: 1,
		UpdatedBy: 1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestPageToolHandler_CreatePage(t *testing.T) {
	svc := &mockPageService{
		createFn: func(_ context.Context, spaceID, userID uint64, _ dto.CreatePageRequest) (*dto.PageResponse, error) {
			if spaceID != 10 {
				t.Errorf("expected spaceID 10, got %d", spaceID)
			}
			if userID != 1 {
				t.Errorf("expected userID 1, got %d", userID)
			}
			return samplePageResponse(), nil
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.CreatePage(context.Background(), nil, CreatePageInput{
		SpaceID: 10,
		Title:   "Test Page",
		Content: "# Hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected page ID 1, got %d", result.ID)
	}
}

func TestPageToolHandler_CreatePage_Error(t *testing.T) {
	svc := &mockPageService{
		createFn: func(context.Context, uint64, uint64, dto.CreatePageRequest) (*dto.PageResponse, error) {
			return nil, errors.New("db error")
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.CreatePage(context.Background(), nil, CreatePageInput{SpaceID: 1, Title: "X"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPageToolHandler_CreatePage_Unauthorized(t *testing.T) {
	svc := &mockPageService{}
	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), denyAuthz(t))

	_, _, err := h.CreatePage(context.Background(), nil, CreatePageInput{SpaceID: 10, Title: "X"})
	if err == nil {
		t.Fatal("expected authorization error, got nil")
	}
}

func TestPageToolHandler_GetPage(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(_ context.Context, pageID uint64) (*dto.PageResponse, error) {
			if pageID != 42 {
				t.Errorf("expected pageID 42, got %d", pageID)
			}
			return samplePageResponse(), nil
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.GetPage(context.Background(), nil, GetPageInput{PageID: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got %q", result.Title)
	}
}

func TestPageToolHandler_GetPage_Unauthorized(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(_ context.Context, _ uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), denyAuthz(t))
	_, _, err := h.GetPage(context.Background(), nil, GetPageInput{PageID: 42})
	if err == nil {
		t.Fatal("expected authorization error, got nil")
	}
}

func TestPageToolHandler_UpdatePage(t *testing.T) {
	newTitle := "Updated Title"
	newContent := "Updated content"

	svc := &mockPageService{
		getByIDFn: func(_ context.Context, _ uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
		updateFn: func(_ context.Context, pageID, userID uint64, req dto.UpdatePageRequest) (*dto.PageResponse, error) {
			if pageID != 5 {
				t.Errorf("expected pageID 5, got %d", pageID)
			}
			if *req.Title != newTitle {
				t.Errorf("expected title %q, got %q", newTitle, *req.Title)
			}
			if *req.Content != newContent {
				t.Errorf("expected content %q, got %q", newContent, *req.Content)
			}
			return samplePageResponse(), nil
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.UpdatePage(context.Background(), nil, UpdatePageInput{
		PageID:  5,
		Title:   &newTitle,
		Content: &newContent,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageToolHandler_DeletePage(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(_ context.Context, _ uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
		deleteFn: func(_ context.Context, pageID uint64) error {
			if pageID != 99 {
				t.Errorf("expected pageID 99, got %d", pageID)
			}
			return nil
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.DeletePage(context.Background(), nil, DeletePageInput{PageID: 99})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success=true")
	}
}

func TestPageToolHandler_DeletePage_Error(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(_ context.Context, _ uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
		deleteFn: func(context.Context, uint64) error {
			return errors.New("not found")
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.DeletePage(context.Background(), nil, DeletePageInput{PageID: 1})
	if err == nil {
		t.Fatal("expected error for delete failure")
	}
}

func TestPageToolHandler_MovePage(t *testing.T) {
	var parentID uint64 = 10
	pos := 2

	svc := &mockPageService{
		getByIDFn: func(_ context.Context, _ uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
		movePageFn: func(_ context.Context, pageID uint64, req dto.MovePageRequest) (*dto.PageResponse, error) {
			if pageID != 5 {
				t.Errorf("expected pageID 5, got %d", pageID)
			}
			if *req.ParentID != parentID {
				t.Errorf("expected parentID 10, got %d", *req.ParentID)
			}
			return samplePageResponse(), nil
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.MovePage(context.Background(), nil, MovePageInput{
		PageID:   5,
		ParentID: &parentID,
		Position: &pos,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageToolHandler_GetPageTree(t *testing.T) {
	svc := &mockPageService{
		getTreeBySpaceFn: func(_ context.Context, spaceID uint64) ([]*dto.PageTreeResponse, error) {
			if spaceID != 10 {
				t.Errorf("expected spaceID 10, got %d", spaceID)
			}
			return []*dto.PageTreeResponse{
				{ID: 1, Title: "Root", Slug: "root", Status: "published", Children: nil},
			}, nil
		},
	}

	h := NewPageToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.GetPageTree(context.Background(), nil, GetPageTreeInput{SpaceID: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Pages) != 1 {
		t.Fatalf("expected 1 tree node, got %d", len(result.Pages))
	}
	if result.Pages[0].Title != "Root" {
		t.Errorf("expected title 'Root', got %q", result.Pages[0].Title)
	}
}
