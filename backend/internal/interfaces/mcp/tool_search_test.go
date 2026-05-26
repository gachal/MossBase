package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/gachal/mossbase/backend/internal/application/dto"
)

func TestSearchToolHandler_Search(t *testing.T) {
	svc := &mockPageService{
		searchFn: func(_ context.Context, spaceID uint64, query string, page, pageSize int) (*dto.SearchResultResponse, error) {
			if spaceID != 5 {
				t.Errorf("expected spaceID 5, got %d", spaceID)
			}
			if query != "golang" {
				t.Errorf("expected query 'golang', got %q", query)
			}
			if page != 1 || pageSize != 20 {
				t.Errorf("expected page=1, pageSize=20, got page=%d, pageSize=%d", page, pageSize)
			}
			return &dto.SearchResultResponse{
				Items:    []dto.SearchResultItem{{ID: 1, Title: "Go Guide", Snippet: "Go is..."}},
				Total:    1,
				Page:     1,
				PageSize: 20,
			}, nil
		},
	}

	h := NewSearchToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.Search(context.Background(), nil, SearchInput{
		SpaceID: 5,
		Query:   "golang",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Title != "Go Guide" {
		t.Errorf("expected title 'Go Guide', got %q", result.Items[0].Title)
	}
}

func TestSearchToolHandler_Search_CustomPagination(t *testing.T) {
	svc := &mockPageService{
		searchFn: func(_ context.Context, _ uint64, _ string, page, pageSize int) (*dto.SearchResultResponse, error) {
			if page != 3 || pageSize != 50 {
				t.Errorf("expected page=3, pageSize=50, got page=%d, pageSize=%d", page, pageSize)
			}
			return &dto.SearchResultResponse{}, nil
		},
	}

	h := NewSearchToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.Search(context.Background(), nil, SearchInput{
		SpaceID:  1,
		Query:    "test",
		Page:     3,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchToolHandler_Search_QueryTooLong(t *testing.T) {
	longQuery := make([]byte, maxQueryLen+1)
	for i := range longQuery {
		longQuery[i] = 'a'
	}

	h := NewSearchToolHandler(nil, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.Search(context.Background(), nil, SearchInput{
		SpaceID: 1,
		Query:   string(longQuery),
	})
	if err == nil {
		t.Fatal("expected error for query exceeding max length")
	}
}

func TestSearchToolHandler_Search_EmptyQuery(t *testing.T) {
	h := NewSearchToolHandler(nil, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.Search(context.Background(), nil, SearchInput{
		SpaceID: 1,
		Query:   "",
	})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearchToolHandler_Search_Error(t *testing.T) {
	svc := &mockPageService{
		searchFn: func(context.Context, uint64, string, int, int) (*dto.SearchResultResponse, error) {
			return nil, errors.New("search failed")
		},
	}

	h := NewSearchToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.Search(context.Background(), nil, SearchInput{SpaceID: 1, Query: "x"})
	if err == nil {
		t.Fatal("expected error for search failure")
	}
}

func TestSearchToolHandler_SemanticSearch(t *testing.T) {
	svc := &mockPageService{
		semanticSearchFn: func(_ context.Context, spaceID uint64, query string, limit int) (*dto.SemanticSearchResponse, error) {
			if spaceID != 5 {
				t.Errorf("expected spaceID 5, got %d", spaceID)
			}
			if query != "architecture patterns" {
				t.Errorf("expected query 'architecture patterns', got %q", query)
			}
			if limit != 10 {
				t.Errorf("expected limit 10, got %d", limit)
			}
			return &dto.SemanticSearchResponse{
				Results: []dto.SemanticSearchItem{
					{PageID: 1, Title: "Patterns", Snippet: "...", Score: 0.95},
				},
				Total: 1,
				Query: "architecture patterns",
			}, nil
		},
	}

	h := NewSearchToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, result, err := h.SemanticSearch(context.Background(), nil, SemanticSearchInput{
		SpaceID: 5,
		Query:   "architecture patterns",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestSearchToolHandler_SemanticSearch_CustomLimit(t *testing.T) {
	svc := &mockPageService{
		semanticSearchFn: func(_ context.Context, _ uint64, _ string, limit int) (*dto.SemanticSearchResponse, error) {
			if limit != 5 {
				t.Errorf("expected limit 5, got %d", limit)
			}
			return &dto.SemanticSearchResponse{}, nil
		},
	}

	h := NewSearchToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.SemanticSearch(context.Background(), nil, SemanticSearchInput{
		SpaceID: 1,
		Query:   "test",
		Limit:   5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchToolHandler_SemanticSearch_Error(t *testing.T) {
	svc := &mockPageService{
		semanticSearchFn: func(context.Context, uint64, string, int) (*dto.SemanticSearchResponse, error) {
			return nil, errors.New("RAG unavailable")
		},
	}

	h := NewSearchToolHandler(svc, NewMCPAuth(nil, 1), allowAllAuthz(t))
	_, _, err := h.SemanticSearch(context.Background(), nil, SemanticSearchInput{SpaceID: 1, Query: "x"})
	if err == nil {
		t.Fatal("expected error for RAG failure")
	}
}
