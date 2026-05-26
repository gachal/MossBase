package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
)

type SpaceMemberRepository interface {
	Create(ctx context.Context, member *entity.SpaceMember) error
	FindBySpaceAndUser(ctx context.Context, spaceID, userID uint64) (*entity.SpaceMember, error)
	FindBySpaceID(ctx context.Context, spaceID uint64) ([]entity.SpaceMember, error)
	FindByUserID(ctx context.Context, userID uint64) ([]entity.SpaceMember, error)
	Delete(ctx context.Context, spaceID, userID uint64) error
	CountAdminsBySpaceID(ctx context.Context, spaceID uint64) (int64, error)
}
