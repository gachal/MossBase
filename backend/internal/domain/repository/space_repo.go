package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type SpaceRepository interface {
	Create(ctx context.Context, space *entity.Space) error
	FindByID(ctx context.Context, id uint64) (*entity.Space, error)
	Update(ctx context.Context, space *entity.Space) error
	Delete(ctx context.Context, id uint64) error
	ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]entity.Space, int64, error)
	ListAll(ctx context.Context, offset, limit int) ([]entity.Space, int64, error)
}
