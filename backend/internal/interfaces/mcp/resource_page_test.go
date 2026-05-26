package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestParsePageURI_Valid(t *testing.T) {
	tests := []struct {
		uri           string
		expectedSpace uint64
		expectedPage  uint64
	}{
		{"mossbase://spaces/10/pages/42", 10, 42},
		{"mossbase://spaces/1/pages/1", 1, 1},
		{"mossbase://spaces/999/pages/12345", 999, 12345},
	}

	for _, tt := range tests {
		spaceID, pageID, err := parsePageURI(tt.uri)
		if err != nil {
			t.Errorf("parsePageURI(%q): unexpected error: %v", tt.uri, err)
			continue
		}
		if spaceID != tt.expectedSpace {
			t.Errorf("parsePageURI(%q): expected spaceID %d, got %d", tt.uri, tt.expectedSpace, spaceID)
		}
		if pageID != tt.expectedPage {
			t.Errorf("parsePageURI(%q): expected pageID %d, got %d", tt.uri, tt.expectedPage, pageID)
		}
	}
}

func TestParsePageURI_Invalid(t *testing.T) {
	tests := []string{
		"mossbase://spaces/10",
		"mossbase://pages/42",
		"http://spaces/10/pages/42",
		"mossbase://spaces/10/pages/abc",
		"mossbase://spaces/abc/pages/42",
		"mossbase://spaces/10/pages/",
		"",
	}

	for _, uri := range tests {
		_, _, err := parsePageURI(uri)
		if err == nil {
			t.Errorf("parsePageURI(%q): expected error, got nil", uri)
		}
	}
}

func TestPageResourceHandler_ReadPageResource(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(_ context.Context, pageID uint64) (*dto.PageResponse, error) {
			if pageID != 42 {
				t.Errorf("expected pageID 42, got %d", pageID)
			}
			return samplePageResponse(), nil
		},
	}

	h := NewPageResourceHandler(svc, allowAllAuthz(t))
	result, err := h.ReadPageResource(context.Background(), &mcpsdk.ReadResourceRequest{
		Params: &mcpsdk.ReadResourceParams{URI: "mossbase://spaces/10/pages/42"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Contents))
	}
	if result.Contents[0].MIMEType != "text/markdown" {
		t.Errorf("expected MIME type 'text/markdown', got %q", result.Contents[0].MIMEType)
	}
	if result.Contents[0].Text == "" {
		t.Error("expected non-empty text content")
	}
}

func TestPageResourceHandler_ReadPageResource_InvalidURI(t *testing.T) {
	svc := &mockPageService{}
	h := NewPageResourceHandler(svc, allowAllAuthz(t))

	_, err := h.ReadPageResource(context.Background(), &mcpsdk.ReadResourceRequest{
		Params: &mcpsdk.ReadResourceParams{URI: "invalid://uri"},
	})
	if err == nil {
		t.Fatal("expected error for invalid URI")
	}
}

func TestPageResourceHandler_ReadPageResource_PageNotFound(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(context.Context, uint64) (*dto.PageResponse, error) {
			return nil, errors.New("page not found")
		},
	}

	h := NewPageResourceHandler(svc, allowAllAuthz(t))
	_, err := h.ReadPageResource(context.Background(), &mcpsdk.ReadResourceRequest{
		Params: &mcpsdk.ReadResourceParams{URI: "mossbase://spaces/10/pages/999"},
	})
	if err == nil {
		t.Fatal("expected error for missing page")
	}
}

func TestPageResourceHandler_ReadPageResource_Unauthorized(t *testing.T) {
	svc := &mockPageService{
		getByIDFn: func(_ context.Context, _ uint64) (*dto.PageResponse, error) {
			return samplePageResponse(), nil
		},
	}

	h := NewPageResourceHandler(svc, denyAuthz(t))
	_, err := h.ReadPageResource(context.Background(), &mcpsdk.ReadResourceRequest{
		Params: &mcpsdk.ReadResourceParams{URI: "mossbase://spaces/10/pages/42"},
	})
	if err == nil {
		t.Fatal("expected authorization error")
	}
}
