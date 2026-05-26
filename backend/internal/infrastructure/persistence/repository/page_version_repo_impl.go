package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

type PageVersionRepoImpl struct {
	db *gorm.DB
}

func NewPageVersionRepository(db *gorm.DB) repository.PageVersionRepository {
	return &PageVersionRepoImpl{db: db}
}

func (r *PageVersionRepoImpl) Create(ctx context.Context, version *entity.PageVersion) error {
	m := model.FromPageVersionEntity(version)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	version.ID = m.ID
	version.CreatedAt = m.CreatedAt
	return nil
}

func (r *PageVersionRepoImpl) FindByPageID(ctx context.Context, pageID uint64, offset, limit int) ([]entity.PageVersion, int64, error) {
	var models []model.PageVersionModel
	var total int64
	db := r.db.WithContext(ctx).Model(&model.PageVersionModel{}).Where("page_id = ?", pageID)
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Order("version_number DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	versions := make([]entity.PageVersion, len(models))
	for i, m := range models {
		versions[i] = *m.ToEntity()
	}
	return versions, total, nil
}

func (r *PageVersionRepoImpl) FindByVersionNumber(ctx context.Context, pageID uint64, versionNumber int) (*entity.PageVersion, error) {
	var m model.PageVersionModel
	if err := r.db.WithContext(ctx).Where("page_id = ? AND version_number = ?", pageID, versionNumber).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *PageVersionRepoImpl) FindLatestVersion(ctx context.Context, pageID uint64) (*entity.PageVersion, error) {
	var m model.PageVersionModel
	if err := r.db.WithContext(ctx).Where("page_id = ?", pageID).Order("version_number DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}
