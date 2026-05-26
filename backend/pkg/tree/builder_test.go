package tree

import (
	"testing"

	"github.com/gachal/mossbase/backend/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestBuildTree(t *testing.T) {
	pages := []entity.Page{
		{ID: 1, Title: "Root 1", ParentID: nil, Position: 0},
		{ID: 2, Title: "Root 2", ParentID: nil, Position: 1},
		{ID: 3, Title: "Child 1.1", ParentID: uint64Ptr(1), Position: 0},
		{ID: 4, Title: "Child 1.2", ParentID: uint64Ptr(1), Position: 1},
		{ID: 5, Title: "Child 1.1.1", ParentID: uint64Ptr(3), Position: 0},
	}

	roots := BuildTree(pages)
	assert.Len(t, roots, 2)
	assert.Equal(t, "Root 1", roots[0].Title)
	assert.Len(t, roots[0].Children, 2)
	assert.Equal(t, "Root 2", roots[1].Title)
	assert.Len(t, roots[1].Children, 0)

	assert.Len(t, roots[0].Children[0].Children, 1)
	assert.Equal(t, "Child 1.1.1", roots[0].Children[0].Children[0].Title)
}

func TestBuildTree_Empty(t *testing.T) {
	roots := BuildTree(nil)
	assert.Len(t, roots, 0)
}

func TestBuildTree_OrphanBecomesRoot(t *testing.T) {
	pages := []entity.Page{
		{ID: 1, Title: "Orphan", ParentID: uint64Ptr(999)},
	}
	roots := BuildTree(pages)
	assert.Len(t, roots, 1)
	assert.Equal(t, "Orphan", roots[0].Title)
}

func TestIsDescendant(t *testing.T) {
	pages := []entity.Page{
		{ID: 1, ParentID: nil},
		{ID: 2, ParentID: uint64Ptr(1)},
		{ID: 3, ParentID: uint64Ptr(2)},
		{ID: 4, ParentID: uint64Ptr(3)},
	}

	tests := []struct {
		name     string
		parentID uint64
		targetID uint64
		want     bool
	}{
		{"direct child", 1, 2, true},
		{"grandchild", 1, 3, true},
		{"great-grandchild", 1, 4, true},
		{"not descendant", 2, 1, false},
		{"self", 1, 1, true},
		{"unrelated", 3, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsDescendant(pages, tt.parentID, tt.targetID))
		})
	}
}

func uint64Ptr(v uint64) *uint64 { return &v }
