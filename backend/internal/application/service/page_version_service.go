package service

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/application/dto"
)

type PageVersionService interface {
	ListVersions(ctx context.Context, pageID uint64, page, pageSize int) ([]dto.PageVersionResponse, int64, error)
	GetVersion(ctx context.Context, pageID uint64, versionNumber int) (*dto.PageVersionResponse, error)
	GetDiff(ctx context.Context, pageID uint64, fromVersion, toVersion int) (*dto.VersionDiffResponse, error)
	RestoreVersion(ctx context.Context, pageID, userID uint64, versionNumber int) (*dto.PageResponse, error)
}
