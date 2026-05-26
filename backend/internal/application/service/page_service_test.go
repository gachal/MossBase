package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

var errNotFound = errors.New("not found")

func ptrStr(s string) *string { return &s }

type mockPageRepo struct {
	pages  map[uint64]*entity.Page
	nextID uint64
}

func newMockPageRepo() *mockPageRepo {
	return &mockPageRepo{pages: make(map[uint64]*entity.Page), nextID: 1}
}

func (m *mockPageRepo) Create(_ context.Context, page *entity.Page) error {
	page.ID = m.nextID
	m.nextID++
	m.pages[page.ID] = page
	return nil
}
func (m *mockPageRepo) FindByID(_ context.Context, id uint64) (*entity.Page, error) {
	p, ok := m.pages[id]
	if !ok { return nil, errNotFound }
	return p, nil
}
func (m *mockPageRepo) FindBySpaceID(_ context.Context, spaceID uint64, offset, limit int) ([]entity.Page, int64, error) {
	return nil, 0, nil
}
func (m *mockPageRepo) FindAllBySpaceID(_ context.Context, spaceID uint64) ([]entity.Page, error) {
	var result []entity.Page
	for _, p := range m.pages {
		if p.SpaceID == spaceID { result = append(result, *p) }
	}
	return result, nil
}
func (m *mockPageRepo) FindByParentID(_ context.Context, spaceID uint64, parentID *uint64) ([]entity.Page, error) {
	return nil, nil
}
func (m *mockPageRepo) Update(_ context.Context, page *entity.Page) error {
	m.pages[page.ID] = page
	return nil
}
func (m *mockPageRepo) Delete(_ context.Context, id uint64) error {
	delete(m.pages, id)
	return nil
}
func (m *mockPageRepo) MaxPositionByParent(_ context.Context, spaceID uint64, parentID *uint64) (int, error) {
	return 0, nil
}
func (m *mockPageRepo) ListAll(_ context.Context, offset, limit int) ([]entity.Page, int64, error) {
	return nil, int64(len(m.pages)), nil
}
func (m *mockPageRepo) Search(_ context.Context, spaceID uint64, query string, offset, limit int) ([]entity.Page, int64, error) {
	return nil, 0, nil
}

type mockPageVersionRepo struct {
	versions []*entity.PageVersion
}

func newMockPageVersionRepo() *mockPageVersionRepo { return &mockPageVersionRepo{} }
func (m *mockPageVersionRepo) Create(_ context.Context, v *entity.PageVersion) error {
	m.versions = append(m.versions, v)
	return nil
}
func (m *mockPageVersionRepo) FindByPageID(_ context.Context, pageID uint64, offset, limit int) ([]entity.PageVersion, int64, error) {
	return nil, 0, nil
}
func (m *mockPageVersionRepo) FindByVersionNumber(_ context.Context, pageID uint64, vn int) (*entity.PageVersion, error) {
	return nil, errNotFound
}
func (m *mockPageVersionRepo) FindLatestVersion(_ context.Context, pageID uint64) (*entity.PageVersion, error) {
	return nil, errNotFound
}

func TestPageService_Create(t *testing.T) {
	svc := NewPageService(newMockPageRepo(), newMockPageVersionRepo(), nil)

	result, err := svc.Create(context.Background(), 1, 100, dto.CreatePageRequest{
		Title: "Test Page", Content: "Hello world",
	})

	assert.NoError(t, err)
	assert.Equal(t, "Test Page", result.Title)
	assert.Equal(t, uint64(1), result.SpaceID)
	assert.Equal(t, 1, result.Version)
}

func TestPageService_GetTreeBySpace(t *testing.T) {
	pageRepo := newMockPageRepo()
	parentID := uint64(1)
	pageRepo.pages[1] = &entity.Page{ID: 1, SpaceID: 10, Title: "Root", ParentID: nil}
	pageRepo.pages[2] = &entity.Page{ID: 2, SpaceID: 10, Title: "Child", ParentID: &parentID}
	pageRepo.nextID = 3

	svc := NewPageService(pageRepo, newMockPageVersionRepo(), nil)
	tree, err := svc.GetTreeBySpace(context.Background(), 10)

	assert.NoError(t, err)
	assert.Len(t, tree, 1)
	assert.Equal(t, "Root", tree[0].Title)
	assert.Len(t, tree[0].Children, 1)
}

func TestPageService_MovePage_CycleDetection(t *testing.T) {
	pageRepo := newMockPageRepo()
	parentID := uint64(1)
	pageRepo.pages[1] = &entity.Page{ID: 1, SpaceID: 10, ParentID: nil}
	pageRepo.pages[2] = &entity.Page{ID: 2, SpaceID: 10, ParentID: &parentID}
	pageRepo.nextID = 3

	svc := NewPageService(pageRepo, newMockPageVersionRepo(), nil)

	newParent := uint64(2)
	_, err := svc.MovePage(context.Background(), 1, dto.MovePageRequest{ParentID: &newParent})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestPageService_Update(t *testing.T) {
	pageRepo := newMockPageRepo()
	pageRepo.pages[1] = &entity.Page{ID: 1, SpaceID: 10, Title: "Old", Version: 1}
	svc := NewPageService(pageRepo, newMockPageVersionRepo(), nil)

	result, err := svc.Update(context.Background(), 1, 100, dto.UpdatePageRequest{
		Title: ptrStr("New Title"), Content: ptrStr("New content"),
	})

	assert.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
	assert.Equal(t, 2, result.Version)
}

func TestPageService_Delete(t *testing.T) {
	pageRepo := newMockPageRepo()
	pageRepo.pages[1] = &entity.Page{ID: 1, SpaceID: 10, Title: "Delete Me"}
	svc := NewPageService(pageRepo, newMockPageVersionRepo(), nil)

	err := svc.Delete(context.Background(), 1)
	assert.NoError(t, err)
	_, ok := pageRepo.pages[1]
	assert.False(t, ok)
}
