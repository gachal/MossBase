package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/pkg/diff"
)

type PageVersionServiceImpl struct {
	pageRepo        repository.PageRepository
	pageVersionRepo repository.PageVersionRepository
}

func NewPageVersionService(
	pageRepo repository.PageRepository,
	pageVersionRepo repository.PageVersionRepository,
) PageVersionService {
	return &PageVersionServiceImpl{pageRepo: pageRepo, pageVersionRepo: pageVersionRepo}
}

func (s *PageVersionServiceImpl) ListVersions(ctx context.Context, pageID uint64, page, pageSize int) ([]dto.PageVersionResponse, int64, error) {
	offset := (page - 1) * pageSize
	versions, total, err := s.pageVersionRepo.FindByPageID(ctx, pageID, offset, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list versions: %w", err)
	}
	result := make([]dto.PageVersionResponse, len(versions))
	for i, v := range versions {
		result[i] = toVersionResponse(&v)
	}
	return result, total, nil
}

func (s *PageVersionServiceImpl) GetVersion(ctx context.Context, pageID uint64, versionNumber int) (*dto.PageVersionResponse, error) {
	version, err := s.pageVersionRepo.FindByVersionNumber(ctx, pageID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("find version: %w", err)
	}
	resp := toVersionResponse(version)
	return &resp, nil
}

func (s *PageVersionServiceImpl) GetDiff(ctx context.Context, pageID uint64, fromVersion, toVersion int) (*dto.VersionDiffResponse, error) {
	v1, err := s.pageVersionRepo.FindByVersionNumber(ctx, pageID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("find from version: %w", err)
	}
	v2, err := s.pageVersionRepo.FindByVersionNumber(ctx, pageID, toVersion)
	if err != nil {
		return nil, fmt.Errorf("find to version: %w", err)
	}
	d := diff.Compute(v1.Content, v2.Content)
	diffJSON, _ := json.Marshal(d)
	return &dto.VersionDiffResponse{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Diff:        string(diffJSON),
	}, nil
}

func (s *PageVersionServiceImpl) RestoreVersion(ctx context.Context, pageID, userID uint64, versionNumber int) (*dto.PageResponse, error) {
	version, err := s.pageVersionRepo.FindByVersionNumber(ctx, pageID, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("find version: %w", err)
	}

	page, err := s.pageRepo.FindByID(ctx, pageID)
	if err != nil {
		return nil, fmt.Errorf("find page: %w", err)
	}

	page.Title = version.Title
	page.Content = version.Content
	page.Version++
	page.UpdatedBy = userID

	if err := s.pageRepo.Update(ctx, page); err != nil {
		return nil, fmt.Errorf("update page: %w", err)
	}

	newVersion := &entity.PageVersion{
		PageID:        page.ID,
		VersionNumber: page.Version,
		Title:         page.Title,
		Content:       page.Content,
		EditedBy:      userID,
	}
	if err := s.pageVersionRepo.Create(ctx, newVersion); err != nil {
		return nil, fmt.Errorf("create restored version: %w", err)
	}

	resp := toPageResponse(page)
	return &resp, nil
}

func toVersionResponse(v *entity.PageVersion) dto.PageVersionResponse {
	return dto.PageVersionResponse{
		ID:            v.ID,
		PageID:        v.PageID,
		VersionNumber: v.VersionNumber,
		Title:         v.Title,
		Content:       v.Content,
		EditedBy:      v.EditedBy,
		CreatedAt:     v.CreatedAt,
	}
}
