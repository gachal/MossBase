package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestParsePromptArgs(t *testing.T) {
	args := map[string]string{
		"page_id":  "42",
		"space_id": "10",
	}

	input, err := parsePromptArgs[SummarizePageInput](args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.PageID != 42 {
		t.Errorf("expected pageID 42, got %d", input.PageID)
	}
	if input.SpaceID != 10 {
		t.Errorf("expected spaceID 10, got %d", input.SpaceID)
	}
}

func TestParsePromptArgs_StringValues(t *testing.T) {
	args := map[string]string{
		"question": "what is Go?",
		"space_id": "5",
	}

	input, err := parsePromptArgs[SearchAndAnswerInput](args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Question != "what is Go?" {
		t.Errorf("expected question 'what is Go?', got %q", input.Question)
	}
	if input.SpaceID != 5 {
		t.Errorf("expected spaceID 5, got %d", input.SpaceID)
	}
}

func TestWikiPromptHandler_SummarizePage(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(_ context.Context, pageID uint64) (*dto.PageResponse, error) {
			if pageID != 42 {
				t.Errorf("expected pageID 42, got %d", pageID)
			}
			return samplePageResponse(), nil
		},
	}

	h := NewWikiPromptHandler(svc, allowAllAuthz(t))
	result, err := h.SummarizePage(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{"page_id": "42", "space_id": "10"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	if result.Messages[0].Role != "user" {
		t.Errorf("expected role 'user', got %q", result.Messages[0].Role)
	}
}

func TestWikiPromptHandler_SummarizePage_PageNotFound(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(context.Context, uint64) (*dto.PageResponse, error) {
			return nil, errors.New("not found")
		},
	}

	h := NewWikiPromptHandler(svc, allowAllAuthz(t))
	_, err := h.SummarizePage(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{"page_id": "99", "space_id": "1"},
		},
	})
	if err == nil {
		t.Fatal("expected error for missing page")
	}
}

func TestWikiPromptHandler_SummarizePage_Unauthorized(t *testing.T) {
	svc := &mockPageService{}
	h := NewWikiPromptHandler(svc, denyAuthz(t))

	_, err := h.SummarizePage(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{"page_id": "42", "space_id": "10"},
		},
	})
	if err == nil {
		t.Fatal("expected authorization error")
	}
}

func TestWikiPromptHandler_SearchAndAnswer(t *testing.T) {
	svc := &mockPageService{
		searchFn: func(_ context.Context, spaceID uint64, query string, page, pageSize int) (*dto.SearchResultResponse, error) {
			if spaceID != 5 {
				t.Errorf("expected spaceID 5, got %d", spaceID)
			}
			return &dto.SearchResultResponse{
				Items: []dto.SearchResultItem{
					{ID: 1, Title: "Go Basics", Snippet: "Go is a compiled language"},
				},
				Total:    1,
				Page:     1,
				PageSize: 10,
			}, nil
		},
	}

	h := NewWikiPromptHandler(svc, allowAllAuthz(t))
	result, err := h.SearchAndAnswer(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{"question": "What is Go?", "space_id": "5"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
}

func TestWikiPromptHandler_SearchAndAnswer_SearchError(t *testing.T) {
	svc := &mockPageService{
		searchFn: func(context.Context, uint64, string, int, int) (*dto.SearchResultResponse, error) {
			return nil, errors.New("search error")
		},
	}

	h := NewWikiPromptHandler(svc, allowAllAuthz(t))
	_, err := h.SearchAndAnswer(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{"question": "test", "space_id": "1"},
		},
	})
	if err == nil {
		t.Fatal("expected error for search failure")
	}
}

func TestWikiPromptHandler_ExplainPage(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(context.Context, uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
	}

	h := NewWikiPromptHandler(svc, allowAllAuthz(t))
	result, err := h.ExplainPage(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{
				"page_id":  "1",
				"space_id": "10",
				"audience": "beginner",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
}

func TestWikiPromptHandler_ExplainPage_DefaultAudience(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(context.Context, uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
	}

	h := NewWikiPromptHandler(svc, allowAllAuthz(t))
	result, err := h.ExplainPage(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{"page_id": "1", "space_id": "10"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tc, ok := result.Messages[0].Content.(*mcpsdk.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	if tc.Text == "" {
		t.Error("expected non-empty prompt text")
	}
}

func TestWikiPromptHandler_CreateFromOutline(t *testing.T) {
	h := NewWikiPromptHandler(nil, allowAllAuthz(t))
	result, err := h.CreateFromOutline(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{
				"space_id": "5",
				"outline":  "# Title\n## Section 1\n## Section 2",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	tc, ok := result.Messages[0].Content.(*mcpsdk.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	if tc.Text == "" {
		t.Error("expected non-empty prompt text")
	}
}

func TestWikiPromptHandler_CreateFromOutline_WithParent(t *testing.T) {
	h := NewWikiPromptHandler(nil, allowAllAuthz(t))

	result, err := h.CreateFromOutline(context.Background(), &mcpsdk.GetPromptRequest{
		Params: &mcpsdk.GetPromptParams{
			Arguments: map[string]string{
				"space_id":  "5",
				"outline":   "# Outline",
				"parent_id": "10",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tc, ok := result.Messages[0].Content.(*mcpsdk.TextContent)
	if !ok {
		t.Fatal("expected TextContent")
	}
	if tc.Text == "" {
		t.Error("expected non-empty prompt text")
	}
}
