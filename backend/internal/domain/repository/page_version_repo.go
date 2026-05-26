package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type PageVersionRepository interface {
	Create(ctx context.Context, version *entity.PageVersion) error
	FindByPageID(ctx context.Context, pageID uint64, offset, limit int) ([]entity.PageVersion, int64, error)
	FindByVersionNumber(ctx context.Context, pageID uint64, versionNumber int) (*entity.PageVersion, error)
	FindLatestVersion(ctx context.Context, pageID uint64) (*entity.PageVersion, error)
}
