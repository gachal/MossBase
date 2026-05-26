package tree

import "github.com/gachal/mossbase/backend/internal/domain/entity"

func BuildTree(pages []entity.Page) []*entity.PageTreeNode {
	nodeMap := make(map[uint64]*entity.PageTreeNode, len(pages))
	for _, p := range pages {
		nodeMap[p.ID] = &entity.PageTreeNode{Page: p, Children: []*entity.PageTreeNode{}}
	}

	var roots []*entity.PageTreeNode
	for _, p := range pages {
		node := nodeMap[p.ID]
		if p.ParentID == nil {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[*p.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	return roots
}

func IsDescendant(pages []entity.Page, parentID, targetID uint64) bool {
	visited := make(map[uint64]bool, len(pages))
	childrenMap := make(map[uint64][]uint64)
	for _, p := range pages {
		if p.ParentID != nil {
			childrenMap[*p.ParentID] = append(childrenMap[*p.ParentID], p.ID)
		}
	}
	var walk func(id uint64) bool
	walk = func(id uint64) bool {
		if id == targetID {
			return true
		}
		if visited[id] {
			return false
		}
		visited[id] = true
		for _, childID := range childrenMap[id] {
			if walk(childID) {
				return true
			}
		}
		return false
	}
	return walk(parentID)
}
