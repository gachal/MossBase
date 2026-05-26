<template>
  <div class="tree-node" :class="{ 'no-select': isDragging }">
    <div
      ref="rowRef"
      class="node-row"
      :class="{
        selected: node.id === selectedId,
        dragging: isDragSource,
        'drop-before': isDropTarget && activeDropZone === 'before',
        'drop-after': isDropTarget && activeDropZone === 'after',
        'drop-inside': isDropTarget && activeDropZone === 'inside',
      }"
      :style="{ paddingLeft: depth * 16 + 12 + 'px' }"
      draggable="true"
      @click="$emit('select', node.id)"
      @contextmenu.prevent="onContextMenu"
      @dragstart="onDragStart"
      @dragend="onDragEnd"
      @dragover="onDragOver"
      @dragleave="onDragLeave"
      @drop="onDrop"
    >
      <el-icon
        v-if="node.children.length > 0"
        class="toggle-icon"
        :class="{ expanded }"
        @click.stop="expanded = !expanded"
      >
        <ArrowRight />
      </el-icon>
      <span v-else class="toggle-placeholder" />
      <el-icon class="doc-icon"><Document /></el-icon>
      <span class="node-title" :title="node.title">{{ node.title }}</span>
    </div>
    <div v-if="expanded && node.children.length > 0">
      <PageTreeNode
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        :selected-id="selectedId"
        :depth="depth + 1"
        @select="$emit('select', $event)"
        @create="$emit('create', $event)"
        @rename="$emit('rename', $event)"
        @delete="$emit('delete', $event)"
        @move="$emit('move', $event)"
      />
    </div>
    <ContextMenu
      v-if="menuVisible"
      :x="menuX"
      :y="menuY"
      :node="node"
      @close="menuVisible = false"
      @create-child="$emit('create', node.id); menuVisible = false"
      @rename="$emit('rename', node); menuVisible = false"
      @delete="$emit('delete', node); menuVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, inject, ref, type Ref } from 'vue'
import { ArrowRight, Document } from '@element-plus/icons-vue'
import type { PageTreeNode } from '@/types/page'
import ContextMenu from './ContextMenu.vue'
import { computeDropPosition, computeMovePayload, isDescendantOf } from '@/utils/tree'
import type { DropPosition } from '@/utils/tree'

interface DragState {
  draggedId: Ref<number | null>
  targetId: Ref<number | null>
  dropPosition: Ref<DropPosition | null>
  tree: Ref<PageTreeNode[]>
}

const props = defineProps<{
  node: PageTreeNode
  selectedId?: number | null
  depth: number
}>()

const emit = defineEmits<{
  select: [pageId: number]
  create: [parentId: number]
  rename: [node: PageTreeNode]
  delete: [node: PageTreeNode]
  move: [payload: { pageId: number; parentId: number | null; position: number }]
}>()

const dragState = inject<DragState>('dragState')!

const rowRef = ref<HTMLElement | null>(null)
const expanded = ref(true)
const menuVisible = ref(false)
const menuX = ref(0)
const menuY = ref(0)
let expandTimer: ReturnType<typeof setTimeout> | null = null

const isDragSource = computed(() => dragState.draggedId.value === props.node.id)
const isDragging = computed(() => dragState.draggedId.value !== null)
const isDropTarget = computed(() => dragState.targetId.value === props.node.id)
const activeDropZone = computed(() => isDropTarget.value ? dragState.dropPosition.value : null)

function onContextMenu(e: MouseEvent) {
  if (isDragging.value) return
  menuX.value = e.clientX
  menuY.value = e.clientY
  menuVisible.value = true
}

function onDragStart(e: DragEvent) {
  dragState.draggedId.value = props.node.id
  e.dataTransfer!.effectAllowed = 'move'
  e.dataTransfer!.setData('text/plain', String(props.node.id))
}

function onDragEnd() {
  dragState.draggedId.value = null
  dragState.targetId.value = null
  dragState.dropPosition.value = null
  clearExpandTimer()
}

function onDragOver(e: DragEvent) {
  e.preventDefault()
  if (dragState.draggedId.value === null) return
  if (dragState.draggedId.value === props.node.id) return

  if (isDescendantOf(dragState.draggedId.value, props.node.id, getRootTree())) {
    e.dataTransfer!.dropEffect = 'none'
    dragState.targetId.value = null
    dragState.dropPosition.value = null
    clearExpandTimer()
    return
  }

  const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
  const pos = computeDropPosition(rect, e.clientY)

  dragState.targetId.value = props.node.id
  dragState.dropPosition.value = pos
  e.dataTransfer!.dropEffect = 'move'

  if (pos === 'inside' && !expanded.value) {
    clearExpandTimer()
    expandTimer = setTimeout(() => { expanded.value = true }, 500)
  } else {
    clearExpandTimer()
  }
}

function onDragLeave() {
  if (dragState.targetId.value === props.node.id) {
    dragState.targetId.value = null
    dragState.dropPosition.value = null
  }
  clearExpandTimer()
}

function onDrop(e: DragEvent) {
  e.preventDefault()
  e.stopPropagation()
  clearExpandTimer()

  const dId = dragState.draggedId.value
  const tId = dragState.targetId.value
  const dPos = dragState.dropPosition.value

  if (dId === null || tId === null || dPos === null) return

  const payload = computeMovePayload(dId, tId, dPos, getRootTree())
  if (payload === null) {
    onDragEnd()
    return
  }

  dragState.draggedId.value = null
  dragState.targetId.value = null
  dragState.dropPosition.value = null

  emit('move', { pageId: dId, parentId: payload.parent_id, position: payload.position })
}

function clearExpandTimer() {
  if (expandTimer !== null) {
    clearTimeout(expandTimer)
    expandTimer = null
  }
}

function getRootTree(): PageTreeNode[] {
  return dragState.tree.value
}
</script>

<style scoped>
.tree-node { position: relative; }
.node-row {
  display: flex; align-items: center; gap: 4px;
  padding: 4px 12px; cursor: pointer; font-size: 13px;
  white-space: nowrap; overflow: hidden;
  position: relative;
  transition: background-color 0.15s;
}
.node-row:hover { background: var(--el-fill-color-light); }
.node-row.selected { background: var(--el-color-primary-light-9); color: var(--el-color-primary); }
.toggle-icon { transition: transform 0.2s; font-size: 12px; flex-shrink: 0; }
.toggle-icon.expanded { transform: rotate(90deg); }
.toggle-placeholder { width: 12px; flex-shrink: 0; }
.doc-icon { font-size: 14px; color: var(--el-text-color-secondary); flex-shrink: 0; }
.node-title { overflow: hidden; text-overflow: ellipsis; }

.node-row.dragging { opacity: 0.4; }
.node-row.drop-inside { background: var(--el-color-primary-light-8); border-radius: 2px; }
.node-row.drop-before::before,
.node-row.drop-after::after {
  content: '';
  position: absolute;
  left: 12px;
  right: 0;
  height: 2px;
  background: var(--el-color-primary);
  border-radius: 1px;
  pointer-events: none;
}
.node-row.drop-before::before { top: -1px; }
.node-row.drop-after::after { bottom: -1px; }

.no-select { user-select: none; -webkit-user-select: none; }
</style>
