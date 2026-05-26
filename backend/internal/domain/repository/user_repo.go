package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint64) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	List(ctx context.Context, offset, limit int) ([]entity.User, int64, error)
	Count(ctx context.Context) (int64, error)
}
