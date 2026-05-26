package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

type SpaceMemberRepoImpl struct {
	db *gorm.DB
}

func NewSpaceMemberRepository(db *gorm.DB) repository.SpaceMemberRepository {
	return &SpaceMemberRepoImpl{db: db}
}

func (r *SpaceMemberRepoImpl) Create(ctx context.Context, member *entity.SpaceMember) error {
	m := model.FromSpaceMemberEntity(member)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	member.ID = m.ID
	member.CreatedAt = m.CreatedAt
	return nil
}

func (r *SpaceMemberRepoImpl) FindBySpaceAndUser(ctx context.Context, spaceID, userID uint64) (*entity.SpaceMember, error) {
	var m model.SpaceMemberModel
	if err := r.db.WithContext(ctx).Where("space_id = ? AND user_id = ?", spaceID, userID).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *SpaceMemberRepoImpl) FindBySpaceID(ctx context.Context, spaceID uint64) ([]entity.SpaceMember, error) {
	var models []model.SpaceMemberModel
	if err := r.db.WithContext(ctx).Where("space_id = ?", spaceID).Find(&models).Error; err != nil {
		return nil, err
	}
	members := make([]entity.SpaceMember, len(models))
	for i, m := range models {
		members[i] = *m.ToEntity()
	}
	return members, nil
}

func (r *SpaceMemberRepoImpl) FindByUserID(ctx context.Context, userID uint64) ([]entity.SpaceMember, error) {
	var models []model.SpaceMemberModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models).Error; err != nil {
		return nil, err
	}
	members := make([]entity.SpaceMember, len(models))
	for i, m := range models {
		members[i] = *m.ToEntity()
	}
	return members, nil
}

func (r *SpaceMemberRepoImpl) Delete(ctx context.Context, spaceID, userID uint64) error {
	return r.db.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", spaceID, userID).
		Delete(&model.SpaceMemberModel{}).Error
}

func (r *SpaceMemberRepoImpl) CountAdminsBySpaceID(ctx context.Context, spaceID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.SpaceMemberModel{}).
		Where("space_id = ? AND role = ?", spaceID, "admin").
		Count(&count).Error
	return count, err
}
