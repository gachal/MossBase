package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/pkg/tree"
)

type PageServiceImpl struct {
	pageRepo        repository.PageRepository
	pageVersionRepo repository.PageVersionRepository
}

func NewPageService(
	pageRepo repository.PageRepository,
	pageVersionRepo repository.PageVersionRepository,
) PageService {
	return &PageServiceImpl{pageRepo: pageRepo, pageVersionRepo: pageVersionRepo}
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
	return &resp, nil
}

func (s *PageServiceImpl) Update(ctx context.Context, pageID, userID uint64, req dto.UpdatePageRequest) (*dto.PageResponse, error) {
	page, err := s.pageRepo.FindByID(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("find page: %w", err)
	}

	if req.Title != "" {
		page.Title = req.Title
		page.Slug = slugify(req.Title)
	}
	if req.Content != "" {
		page.Content = req.Content
		if req.ContentHTML != "" {
			page.ContentHTML = req.ContentHTML
		} else {
			page.ContentHTML = req.Content
		}
	}
	page.Version++
	page.UpdatedBy = userID
	page.UpdatedAt = time.Now()

	if err := s.pageRepo.Update(ctx, page); err != nil {
		return nil, fmt.Errorf("update page: %w", err)
	}

	version := &entity.PageVersion{
		PageID:        page.ID,
		VersionNumber: page.Version,
		Title:         page.Title,
		Content:       page.Content,
		EditedBy:      userID,
	}
	if err := s.pageVersionRepo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("create version: %w", err)
	}

	resp := toPageResponse(page)
	return &resp, nil
}

func (s *PageServiceImpl) Delete(ctx context.Context, pageID uint64) error {
	return s.pageRepo.Delete(ctx, pageID)
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
