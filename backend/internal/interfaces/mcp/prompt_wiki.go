package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/gachal/mossbase/backend/internal/application/service"
)

// WikiPromptHandler provides MCP prompt handlers for wiki operations.
type WikiPromptHandler struct {
	pageSvc service.PageService
	authz   *spaceAuthorizer
}

// NewWikiPromptHandler creates a new WikiPromptHandler.
func NewWikiPromptHandler(pageSvc service.PageService, authz *spaceAuthorizer) *WikiPromptHandler {
	return &WikiPromptHandler{pageSvc: pageSvc, authz: authz}
}

// SummarizePageInput defines the arguments for the summarize_page prompt.
type SummarizePageInput struct {
	PageID  uint64 `json:"page_id"`
	SpaceID uint64 `json:"space_id"`
}

// SearchAndAnswerInput defines the arguments for the search_and_answer prompt.
type SearchAndAnswerInput struct {
	Question string `json:"question"`
	SpaceID  uint64 `json:"space_id"`
}

// ExplainPageInput defines the arguments for the explain_page prompt.
type ExplainPageInput struct {
	PageID   uint64 `json:"page_id"`
	SpaceID  uint64 `json:"space_id"`
	Audience string `json:"audience"`
}

// CreateFromOutlineInput defines the arguments for the create_from_outline prompt.
type CreateFromOutlineInput struct {
	SpaceID  uint64  `json:"space_id"`
	Outline  string  `json:"outline"`
	ParentID *uint64 `json:"parent_id,omitempty"`
}

// SummarizePage fetches a page and returns a prompt asking for a summary.
func (h *WikiPromptHandler) SummarizePage(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	input, err := parsePromptArgs[SummarizePageInput](req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("summarize_page: %w", err)
	}

	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, err
	}

	resp, err := h.pageSvc.GetByID(ctx, input.PageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page %d: %w", input.PageID, err)
	}

	promptText := fmt.Sprintf(
		"请总结以下知识库页面的核心内容，用简洁的中文列出主要要点：\n\n## %s\n\n%s",
		resp.Title, resp.Content,
	)

	return &mcpsdk.GetPromptResult{
		Description: "总结指定知识库页面的核心内容",
		Messages: []*mcpsdk.PromptMessage{
			{
				Role:    "user",
				Content: &mcpsdk.TextContent{Text: promptText},
			},
		},
	}, nil
}

// SearchAndAnswer searches the knowledge base and returns a prompt to answer a question.
func (h *WikiPromptHandler) SearchAndAnswer(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	input, err := parsePromptArgs[SearchAndAnswerInput](req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("search_and_answer: %w", err)
	}

	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, err
	}

	searchResult, sErr := h.pageSvc.Search(ctx, input.SpaceID, input.Question, 1, 10)
	if sErr != nil {
		return nil, fmt.Errorf("failed to search in space %d: %w", input.SpaceID, sErr)
	}

	var sb strings.Builder
	sb.WriteString("以下是从知识库中搜索到的相关内容：\n\n")
	for i, item := range searchResult.Items {
		sb.WriteString(fmt.Sprintf("### 结果 %d: %s\n%s\n\n", i+1, item.Title, item.Snippet))
	}

	sb.WriteString(fmt.Sprintf("---\n\n请根据以上搜索结果回答问题：\n%s", input.Question))

	return &mcpsdk.GetPromptResult{
		Description: "在知识库中搜索相关内容并回答问题",
		Messages: []*mcpsdk.PromptMessage{
			{
				Role:    "user",
				Content: &mcpsdk.TextContent{Text: sb.String()},
			},
		},
	}, nil
}

// ExplainPage fetches a page and returns a prompt asking for an explanation.
func (h *WikiPromptHandler) ExplainPage(ctx context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	input, err := parsePromptArgs[ExplainPageInput](req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("explain_page: %w", err)
	}

	if err := h.authz.checkRead(ctx, input.SpaceID); err != nil {
		return nil, err
	}

	audience := input.Audience
	if audience == "" {
		audience = "beginner"
	}

	resp, err := h.pageSvc.GetByID(ctx, input.PageID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page %d: %w", input.PageID, err)
	}

	audienceDesc := map[string]string{
		"beginner":     "初学者（没有相关背景知识）",
		"intermediate": "中级读者（有一定基础）",
		"expert":       "专家（深入了解领域知识）",
	}
	audienceText, ok := audienceDesc[audience]
	if !ok {
		audienceText = audience
	}

	promptText := fmt.Sprintf(
		"请用通俗易懂的语言向%s解释以下知识库页面的内容：\n\n## %s\n\n%s",
		audienceText, resp.Title, resp.Content,
	)

	return &mcpsdk.GetPromptResult{
		Description: "用通俗语言解释知识库页面内容",
		Messages: []*mcpsdk.PromptMessage{
			{
				Role:    "user",
				Content: &mcpsdk.TextContent{Text: promptText},
			},
		},
	}, nil
}

// CreateFromOutline returns a prompt that expands a Markdown outline into full content.
func (h *WikiPromptHandler) CreateFromOutline(_ context.Context, req *mcpsdk.GetPromptRequest) (*mcpsdk.GetPromptResult, error) {
	input, err := parsePromptArgs[CreateFromOutlineInput](req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("create_from_outline: %w", err)
	}

	var parentHint string
	if input.ParentID != nil {
		parentHint = fmt.Sprintf("\n父页面 ID: %d", *input.ParentID)
	}

	promptText := fmt.Sprintf(
		"请根据以下 Markdown 大纲，为知识空间（ID: %d）创建结构化的页面内容。"+
			"每个标题应展开为完整的段落，保持逻辑清晰、内容专业。%s\n\n%s",
		input.SpaceID, parentHint, input.Outline,
	)

	return &mcpsdk.GetPromptResult{
		Description: "根据大纲创建结构化页面内容",
		Messages: []*mcpsdk.PromptMessage{
			{
				Role:    "user",
				Content: &mcpsdk.TextContent{Text: promptText},
			},
		},
	}, nil
}

// parsePromptArgs deserializes the prompt arguments map into a typed struct.
func parsePromptArgs[T any](args map[string]string) (T, error) {
	var zero T

	m := make(map[string]any, len(args))
	for k, v := range args {
		if num, err := strconv.ParseUint(v, 10, 64); err == nil {
			m[k] = num
		} else {
			m[k] = v
		}
	}

	data, err := json.Marshal(m)
	if err != nil {
		return zero, fmt.Errorf("marshal args: %w", err)
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, fmt.Errorf("unmarshal args: %w", err)
	}
	return result, nil
}
