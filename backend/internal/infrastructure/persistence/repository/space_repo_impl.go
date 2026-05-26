package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

type SpaceRepoImpl struct {
	db *gorm.DB
}

func NewSpaceRepository(db *gorm.DB) repository.SpaceRepository {
	return &SpaceRepoImpl{db: db}
}

func (r *SpaceRepoImpl) Create(ctx context.Context, space *entity.Space) error {
	m := model.FromSpaceEntity(space)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	space.ID = m.ID
	space.CreatedAt = m.CreatedAt
	space.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *SpaceRepoImpl) FindByID(ctx context.Context, id uint64) (*entity.Space, error) {
	var m model.SpaceModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *SpaceRepoImpl) Update(ctx context.Context, space *entity.Space) error {
	m := model.FromSpaceEntity(space)
	return r.db.WithContext(ctx).Model(&model.SpaceModel{}).Where("id = ?", space.ID).Updates(m).Error
}

func (r *SpaceRepoImpl) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.SpaceModel{}, id).Error
}

func (r *SpaceRepoImpl) ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]entity.Space, int64, error) {
	var models []model.SpaceModel
	var total int64
	db := r.db.WithContext(ctx).Model(&model.SpaceModel{}).
		Joins("JOIN space_members ON space_members.space_id = spaces.id").
		Where("space_members.user_id = ?", userID)
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	spaces := make([]entity.Space, len(models))
	for i, m := range models {
		spaces[i] = *m.ToEntity()
	}
	return spaces, total, nil
}

func (r *SpaceRepoImpl) ListAll(ctx context.Context, offset, limit int) ([]entity.Space, int64, error) {
	var models []model.SpaceModel
	var total int64
	db := r.db.WithContext(ctx).Model(&model.SpaceModel{})
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	spaces := make([]entity.Space, len(models))
	for i, m := range models {
		spaces[i] = *m.ToEntity()
	}
	return spaces, total, nil
}
