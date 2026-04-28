import { describe, it, expect } from 'vitest'
import { buildTree, flattenTree, findNode, sortByPosition } from '../tree'
import type { PageTreeNode } from '@/types/page'

const makeNode = (overrides: Partial<PageTreeNode> & { id: number }): PageTreeNode => ({
  space_id: 1,
  parent_id: null,
  title: `Page ${overrides.id}`,
  slug: `page-${overrides.id}`,
  content: '',
  content_html: '',
  position: 0,
  status: 'draft',
  version: 1,
  created_by: 1,
  updated_by: 1,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  children: [],
  ...overrides,
})

describe('buildTree', () => {
  it('builds tree from flat list', () => {
    const pages = [
      makeNode({ id: 1, parent_id: null, position: 0 }),
      makeNode({ id: 2, parent_id: 1, position: 0 }),
      makeNode({ id: 3, parent_id: 1, position: 1 }),
      makeNode({ id: 4, parent_id: null, position: 1 }),
    ]
    const tree = buildTree(pages)
    expect(tree).toHaveLength(2)
    expect(tree[0].children).toHaveLength(2)
    expect(tree[1].children).toHaveLength(0)
  })

  it('handles empty list', () => {
    expect(buildTree([])).toHaveLength(0)
  })

  it('handles nested children', () => {
    const pages = [
      makeNode({ id: 1, parent_id: null }),
      makeNode({ id: 2, parent_id: 1 }),
      makeNode({ id: 3, parent_id: 2 }),
    ]
    const tree = buildTree(pages)
    expect(tree[0].children[0].children).toHaveLength(1)
  })
})

describe('flattenTree', () => {
  it('flattens nested tree', () => {
    const tree = [
      makeNode({ id: 1, children: [
        makeNode({ id: 2, children: [] }),
      ]}),
    ]
    expect(flattenTree(tree)).toHaveLength(2)
  })
})

describe('findNode', () => {
  it('finds node by id', () => {
    const tree = [
      makeNode({ id: 1, children: [
        makeNode({ id: 2, children: [] }),
      ]}),
    ]
    expect(findNode(tree, 2)?.id).toBe(2)
  })

  it('returns null for missing id', () => {
    expect(findNode([makeNode({ id: 1 })], 99)).toBeNull()
  })
})

describe('sortByPosition', () => {
  it('sorts by position', () => {
    const nodes = [
      makeNode({ id: 1, position: 2 }),
      makeNode({ id: 2, position: 0 }),
      makeNode({ id: 3, position: 1 }),
    ]
    const sorted = sortByPosition(nodes)
    expect(sorted.map(n => n.id)).toEqual([2, 3, 1])
  })
})
