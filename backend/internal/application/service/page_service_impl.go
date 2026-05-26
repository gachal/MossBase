package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/pkg/tree"
)

type PageServiceImpl struct {
	pageRepo        repository.PageRepository
	pageVersionRepo repository.PageVersionRepository
	ragClient       RAGIndexer
}

func NewPageService(
	pageRepo repository.PageRepository,
	pageVersionRepo repository.PageVersionRepository,
	ragClient RAGIndexer,
) PageService {
	return &PageServiceImpl{
		pageRepo:        pageRepo,
		pageVersionRepo: pageVersionRepo,
		ragClient:       ragClient,
	}
}

func (s *PageServiceImpl) Create(ctx context.Context, spaceID, userID uint64, req dto.CreatePageRequest) (*dto.PageResponse, error) {
	pos := 0
	if req.Position != nil {
		pos = *req.Position
	} else {
		max, err := s.pageRepo.MaxPositionByParent(ctx, spaceID, req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("get max position: %w", err)
		}
		pos = max + 1
	}

	page := &entity.Page{
		SpaceID:   spaceID,
		ParentID:  req.ParentID,
		Title:     req.Title,
		Slug:      slugify(req.Title),
		Content:   req.Content,
		Position:  pos,
		Status:    entity.PageStatusDraft,
		Version:   1,
		CreatedBy: userID,
		UpdatedBy: userID,
	}

	if err := s.pageRepo.Create(ctx, page); err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}

	version := &entity.PageVersion{
		PageID:        page.ID,
		VersionNumber: 1,
		Title:         page.Title,
		Content:       page.Content,
		EditedBy:      userID,
	}
	if err := s.pageVersionRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create initial version: %w", err)
	}

	resp := toPageResponse(page)

	// Sync with RAG service asynchronously
	if s.ragClient != nil {
		rc := s.ragClient
		docID := fmt.Sprintf("page-%d-%d", spaceID, page.ID)
		title := page.Title
		content := page.Content
		go func() {
			defer func() {
				if r := recover(); r != nil {
					zap.L().Warn("RAG IndexDocument goroutine panicked", zap.Any("recover", r))
				}
			}()
			_ = rc.IndexDocument(context.Background(), docID, spaceID, title, content)
		}()
	}

	return &resp, nil
}

func (s *PageServiceImpl) Update(ctx context.Context, pageID, userID uint64, req dto.UpdatePageRequest) (*dto.PageResponse, error) {
	page, err := s.pageRepo.FindByID(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("find page: %w", err)
	}

	updated := &entity.Page{
		ID:          page.ID,
		SpaceID:     page.SpaceID,
		ParentID:    page.ParentID,
		Title:       page.Title,
		Slug:        page.Slug,
		Content:     page.Content,
		ContentHTML: page.ContentHTML,
		Position:    page.Position,
		Status:      page.Status,
		Version:     page.Version + 1,
		CreatedBy:   page.CreatedBy,
		UpdatedBy:   userID,
		UpdatedAt:   time.Now(),
	}

	if req.Title != nil {
		updated.Title = *req.Title
		updated.Slug = slugify(*req.Title)
	}
	if req.Content != nil {
		updated.Content = *req.Content
		if req.ContentHTML != nil && *req.ContentHTML != "" {
			updated.ContentHTML = *req.ContentHTML
		} else {
			updated.ContentHTML = *req.Content
		}
	}

	if err := s.pageRepo.Update(ctx, updated); err != nil {
		return nil, fmt.Errorf("update page: %w", err)
	}

	version := &entity.PageVersion{
		PageID:        updated.ID,
		VersionNumber: updated.Version,
		Title:         updated.Title,
		Content:       updated.Content,
		EditedBy:      userID,
	}
	if err := s.pageVersionRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create version: %w", err)
	}

	resp := toPageResponse(updated)

	// Sync with RAG service asynchronously (re-index)
	if s.ragClient != nil {
		rc := s.ragClient
		docID := fmt.Sprintf("page-%d-%d", updated.SpaceID, updated.ID)
		title := updated.Title
		content := updated.Content
		go func() {
			defer func() {
				if r := recover(); r != nil {
					zap.L().Warn("RAG IndexDocument goroutine panicked", zap.Any("recover", r))
				}
			}()
			_ = rc.IndexDocument(context.Background(), docID, updated.SpaceID, title, content)
		}()
	}

	return &resp, nil
}

func (s *PageServiceImpl) Delete(ctx context.Context, pageID uint64) error {
	// Fetch page to get spaceID for RAG cleanup before deletion
	var spaceID uint64
	if s.ragClient != nil {
		page, err := s.pageRepo.FindByID(ctx, pageID)
		if err != nil {
			zap.L().Warn("failed to fetch page for RAG cleanup",
				zap.Uint64("page_id", pageID),
				zap.Error(err),
			)
		} else {
			spaceID = page.SpaceID
		}
	}

	err := s.pageRepo.Delete(ctx, pageID)
	if err != nil {
		return err
	}

	// Sync with RAG service asynchronously
	if s.ragClient != nil && spaceID > 0 {
		rc := s.ragClient
		docID := fmt.Sprintf("page-%d-%d", spaceID, pageID)
		sid := spaceID
		go func() {
			defer func() {
				if r := recover(); r != nil {
					zap.L().Warn("RAG DeleteDocument goroutine panicked", zap.Any("recover", r))
				}
			}()
			_ = rc.DeleteDocument(context.Background(), docID, sid)
		}()
	}

	return nil
}

func (s *PageServiceImpl) GetByID(ctx context.Context, pageID uint64) (*dto.PageResponse, error) {
	page, err := s.pageRepo.FindByID(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("find page: %w", err)
	}
	resp := toPageResponse(page)
	return &resp, nil
}

func (s *PageServiceImpl) GetTreeBySpace(ctx context.Context, spaceID uint64) ([]*dto.PageTreeResponse, error) {
	pages, err := s.pageRepo.FindAllBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("load pages: %w", err)
	}

	treeNodes := tree.BuildTree(pages)
	return toPageTreeResponses(treeNodes), nil
}

func (s *PageServiceImpl) MovePage(ctx context.Context, pageID uint64, req dto.MovePageRequest) (*dto.PageResponse, error) {
	page, err := s.pageRepo.FindByID(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("find page: %w", err)
	}

	if req.ParentID != nil {
		if *req.ParentID == pageID {
			return nil, fmt.Errorf("page cannot be its own parent")
		}
		if *req.ParentID != 0 {
			allPages, err := s.pageRepo.FindAllBySpaceID(ctx, page.SpaceID)
			if err != nil {
				return nil, fmt.Errorf("load pages for cycle check: %w", err)
			}
			if tree.IsDescendant(allPages, pageID, *req.ParentID) {
				return nil, fmt.Errorf("circular reference detected")
			}
		}
	}

	if req.ParentID != nil {
		if *req.ParentID == 0 {
			page.ParentID = nil
		} else {
			page.ParentID = req.ParentID
		}
	}
	if req.Position != nil {
		page.Position = *req.Position
	}

	if err := s.pageRepo.Update(ctx, page); err != nil {
		return nil, fmt.Errorf("move page: %w", err)
	}

	resp := toPageResponse(page)
	return &resp, nil
}

func toPageResponse(p *entity.Page) dto.PageResponse {
	return dto.PageResponse{
		ID:          p.ID,
		SpaceID:     p.SpaceID,
		ParentID:    p.ParentID,
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		ContentHTML: p.ContentHTML,
		Position:    p.Position,
		Status:      string(p.Status),
		Version:     p.Version,
		CreatedBy:   p.CreatedBy,
		UpdatedBy:   p.UpdatedBy,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toPageTreeResponses(nodes []*entity.PageTreeNode) []*dto.PageTreeResponse {
	result := make([]*dto.PageTreeResponse, len(nodes))
	for i, n := range nodes {
		result[i] = &dto.PageTreeResponse{
			ID:        n.ID,
			SpaceID:   n.SpaceID,
			ParentID:  n.ParentID,
			Title:     n.Title,
			Slug:      n.Slug,
			Position:  n.Position,
			Status:    string(n.Status),
			Version:   n.Version,
			UpdatedAt: n.UpdatedAt,
			Children:  toPageTreeResponses(n.Children),
		}
	}
	return result
}

func slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	result := make([]byte, 0, len(s))
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		}
	}
	return string(result)
}

func (s *PageServiceImpl) Search(ctx context.Context, spaceID uint64, query string, page, pageSize int) (*dto.SearchResultResponse, error) {
	offset := (page - 1) * pageSize
	pages, total, err := s.pageRepo.Search(ctx, spaceID, query, offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("search pages: %w", err)
	}

	items := make([]dto.SearchResultItem, len(pages))
	for i, p := range pages {
		items[i] = dto.SearchResultItem{
			ID:        p.ID,
			Title:     p.Title,
			Snippet:   extractSnippet(p.Content, query, 120),
			UpdatedAt: p.UpdatedAt,
		}
	}

	return &dto.SearchResultResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *PageServiceImpl) SemanticSearch(ctx context.Context, spaceID uint64, query string, limit int) (*dto.SemanticSearchResponse, error) {
	if s.ragClient == nil {
		return nil, fmt.Errorf("RAG service not enabled")
	}

	ragResp, err := s.ragClient.SemanticSearch(ctx, spaceID, query, limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	items := make([]dto.SemanticSearchItem, 0, len(ragResp.Results))
	for _, r := range ragResp.Results {
		var spaceIDParsed, pageID uint64
		n, err := fmt.Sscanf(r.DocumentID, "page-%d-%d", &spaceIDParsed, &pageID)
		if err != nil || n != 2 {
			zap.L().Warn("failed to parse document ID",
				zap.String("document_id", r.DocumentID),
				zap.Error(err),
			)
			continue
		}

		items = append(items, dto.SemanticSearchItem{
			PageID:  pageID,
			Title:   r.Title,
			Snippet: r.Content,
			Score:   float64(r.Score),
		})
	}

	return &dto.SemanticSearchResponse{
		Results: items,
		Total:   len(items),
		Query:   query,
	}, nil
}

func extractSnippet(content, query string, maxLen int) string {
	if content == "" {
		return ""
	}
	idx := strings.Index(strings.ToLower(content), strings.ToLower(query))
	if idx == -1 {
		if len(content) > maxLen {
			return content[:maxLen] + "..."
		}
		return content
	}
	start := idx - maxLen/3
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(content) {
		end = len(content)
	}
	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}
	return snippet
}
