import type { Page, PageTreeNode } from '@/types/page'

export function buildTree(pages: Page[]): PageTreeNode[] {
  const nodeMap = new Map<number, PageTreeNode>()
  for (const p of pages) {
    nodeMap.set(p.id, { ...p, children: [] })
  }

  const roots: PageTreeNode[] = []
  for (const p of pages) {
    const node = nodeMap.get(p.id)!
    if (p.parent_id === null || !nodeMap.has(p.parent_id)) {
      roots.push(node)
    } else {
      const parent = nodeMap.get(p.parent_id)!
      parent.children.push(node)
    }
  }
  return roots
}

export function flattenTree(tree: PageTreeNode[]): Page[] {
  const result: Page[] = []
  const walk = (nodes: PageTreeNode[]) => {
    for (const node of nodes) {
      const { children, ...page } = node
      result.push(page)
      if (children.length > 0) walk(children)
    }
  }
  walk(tree)
  return result
}

export function findNode(tree: PageTreeNode[], id: number): PageTreeNode | null {
  for (const node of tree) {
    if (node.id === id) return node
    const found = findNode(node.children, id)
    if (found) return found
  }
  return null
}

export function sortByPosition(nodes: PageTreeNode[]): PageTreeNode[] {
  return [...nodes].sort((a, b) => a.position - b.position)
}

export function findParent(tree: PageTreeNode[], nodeId: number): PageTreeNode | null {
  for (const node of tree) {
    if (node.children.some(c => c.id === nodeId)) return node
    const found = findParent(node.children, nodeId)
    if (found) return found
  }
  return null
}

export function isDescendantOf(ancestorId: number, nodeId: number, tree: PageTreeNode[]): boolean {
  const ancestor = findNode(tree, ancestorId)
  if (!ancestor) return false
  return findNode(ancestor.children, nodeId) !== null
}

export type DropPosition = 'before' | 'inside' | 'after'

export function computeDropPosition(rect: DOMRect, mouseY: number): DropPosition {
  const relativeY = mouseY - rect.top
  const threshold = rect.height * 0.25
  if (relativeY < threshold) return 'before'
  if (relativeY > rect.height - threshold) return 'after'
  return 'inside'
}

export interface MovePayload {
  parent_id: number | null
  position: number
}

export function computeMovePayload(
  draggedId: number,
  targetId: number,
  dropPosition: DropPosition,
  tree: PageTreeNode[]
): MovePayload | null {
  if (dropPosition === 'inside') {
    const target = findNode(tree, targetId)
    if (!target) return null
    if (target.children.length > 0 && target.children[target.children.length - 1].id === draggedId) {
      return null
    }
    return { parent_id: targetId, position: target.children.length }
  }

  const targetParent = findParent(tree, targetId)
  const siblings = targetParent ? targetParent.children : tree
  const targetIndex = siblings.findIndex(n => n.id === targetId)
  if (targetIndex === -1) return null

  const insertIndex = dropPosition === 'before' ? targetIndex : targetIndex + 1
  if (siblings[insertIndex]?.id === draggedId) return null

  return {
    parent_id: targetParent?.id ?? null,
    position: insertIndex,
  }
}
