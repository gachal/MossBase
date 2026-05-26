package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type PageRepository interface {
	Create(ctx context.Context, page *entity.Page) error
	FindByID(ctx context.Context, id uint64) (*entity.Page, error)
	FindBySpaceID(ctx context.Context, spaceID uint64, offset, limit int) ([]entity.Page, int64, error)
	FindAllBySpaceID(ctx context.Context, spaceID uint64) ([]entity.Page, error)
	FindByParentID(ctx context.Context, spaceID uint64, parentID *uint64) ([]entity.Page, error)
	Update(ctx context.Context, page *entity.Page) error
	Delete(ctx context.Context, id uint64) error
	MaxPositionByParent(ctx context.Context, spaceID uint64, parentID *uint64) (int, error)
	ListAll(ctx context.Context, offset, limit int) ([]entity.Page, int64, error)
	Search(ctx context.Context, spaceID uint64, query string, offset, limit int) ([]entity.Page, int64, error)
}
