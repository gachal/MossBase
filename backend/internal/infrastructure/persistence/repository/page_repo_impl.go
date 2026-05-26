package repository

import (
	"context"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/gachal/mossbase/backend/internal/domain/repository"
	"github.com/gachal/mossbase/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

type PageRepoImpl struct {
	db *gorm.DB
}

func NewPageRepository(db *gorm.DB) repository.PageRepository {
	return &PageRepoImpl{db: db}
}

func (r *PageRepoImpl) Create(ctx context.Context, page *entity.Page) error {
	m := model.FromPageEntity(page)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	page.ID = m.ID
	page.CreatedAt = m.CreatedAt
	page.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *PageRepoImpl) FindByID(ctx context.Context, id uint64) (*entity.Page, error) {
	var m model.PageModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *PageRepoImpl) FindBySpaceID(ctx context.Context, spaceID uint64, offset, limit int) ([]entity.Page, int64, error) {
	var models []model.PageModel
	var total int64
	db := r.db.WithContext(ctx).Model(&model.PageModel{}).Where("space_id = ?", spaceID)
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Order("position ASC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	pages := make([]entity.Page, len(models))
	for i, m := range models {
		pages[i] = *m.ToEntity()
	}
	return pages, total, nil
}

func (r *PageRepoImpl) FindAllBySpaceID(ctx context.Context, spaceID uint64) ([]entity.Page, error) {
	var models []model.PageModel
	if err := r.db.WithContext(ctx).Where("space_id = ?", spaceID).Order("position ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	pages := make([]entity.Page, len(models))
	for i, m := range models {
		pages[i] = *m.ToEntity()
	}
	return pages, nil
}

func (r *PageRepoImpl) FindByParentID(ctx context.Context, spaceID uint64, parentID *uint64) ([]entity.Page, error) {
	var models []model.PageModel
	db := r.db.WithContext(ctx).Where("space_id = ?", spaceID)
	if parentID == nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", *parentID)
	}
	if err := db.Order("position ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	pages := make([]entity.Page, len(models))
	for i, m := range models {
		pages[i] = *m.ToEntity()
	}
	return pages, nil
}

func (r *PageRepoImpl) Update(ctx context.Context, page *entity.Page) error {
	m := model.FromPageEntity(page)
	return r.db.WithContext(ctx).Model(&model.PageModel{}).Where("id = ?", page.ID).
		Select("title", "slug", "content", "content_html", "parent_id", "position", "status", "version", "updated_by", "updated_at").
		Updates(m).Error
}

func (r *PageRepoImpl) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.PageModel{}, id).Error
}

func (r *PageRepoImpl) MaxPositionByParent(ctx context.Context, spaceID uint64, parentID *uint64) (int, error) {
	var maxPos *int
	db := r.db.WithContext(ctx).Model(&model.PageModel{}).Where("space_id = ?", spaceID)
	if parentID == nil {
		db = db.Where("parent_id IS NULL")
	} else {
		db = db.Where("parent_id = ?", *parentID)
	}
	db.Select("MAX(position)").Scan(&maxPos)
	if maxPos == nil {
		return 0, nil
	}
	return *maxPos, nil
}

func (r *PageRepoImpl) ListAll(ctx context.Context, offset, limit int) ([]entity.Page, int64, error) {
	var models []model.PageModel
	var total int64
	db := r.db.WithContext(ctx).Model(&model.PageModel{})
	db.Count(&total)
	if err := db.Offset(offset).Limit(limit).Order("updated_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	pages := make([]entity.Page, len(models))
	for i, m := range models {
		pages[i] = *m.ToEntity()
	}
	return pages, total, nil
}

func (r *PageRepoImpl) Search(ctx context.Context, spaceID uint64, query string, offset, limit int) ([]entity.Page, int64, error) {
	var total int64
	r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM pages WHERE space_id = ? AND MATCH(title, content) AGAINST(? IN BOOLEAN MODE) AND deleted_at IS NULL",
		spaceID, query,
	).Scan(&total)

	var models []model.PageModel
	r.db.WithContext(ctx).Raw(
		"SELECT *, MATCH(title, content) AGAINST(? IN BOOLEAN MODE) AS relevance "+
			"FROM pages WHERE space_id = ? AND MATCH(title, content) AGAINST(? IN BOOLEAN MODE) AND deleted_at IS NULL "+
			"ORDER BY relevance DESC LIMIT ? OFFSET ?",
		query, spaceID, query, limit, offset,
	).Scan(&models)

	pages := make([]entity.Page, len(models))
	for i, m := range models {
		pages[i] = *m.ToEntity()
	}
	return pages, total, nil
}
