package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

type UserRepoImpl struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &UserRepoImpl{db: db}
}

func (r *UserRepoImpl) Create(ctx context.Context, user *entity.User) error {
	m := model.FromUserEntity(user)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *UserRepoImpl) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *UserRepoImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var m model.UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *UserRepoImpl) Update(ctx context.Context, user *entity.User) error {
	m := model.FromUserEntity(user)
	return r.db.WithContext(ctx).Model(&model.UserModel{}).Where("id = ?", user.ID).Updates(m).Error
}

func (r *UserRepoImpl) List(ctx context.Context, offset, limit int) ([]entity.User, int64, error) {
	var models []model.UserModel
	var total int64
	db := r.db.WithContext(ctx).Model(&model.UserModel{})
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	users := make([]entity.User, len(models))
	for i, m := range models {
		users[i] = *m.ToEntity()
	}
	return users, total, nil
}

func (r *UserRepoImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.UserModel{}).Count(&count).Error
	return count, err
}
