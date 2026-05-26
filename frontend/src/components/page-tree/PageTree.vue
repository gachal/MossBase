<template>
  <div class="page-tree">
    <div class="tree-header">
      <span class="tree-title">页面</span>
      <el-button size="small" text @click="$emit('create', null)">
        <el-icon><Plus /></el-icon>
      </el-button>
    </div>
    <div class="tree-search">
      <el-input
        v-model="searchText"
        placeholder="搜索页面..."
        size="small"
        clearable
        :prefix-icon="Search"
        @input="handleSearchInput"
        @clear="handleSearchClear"
      />
      <el-radio-group v-model="localSearchMode" size="small" style="margin-top: 4px">
        <el-radio-button value="keyword">关键词</el-radio-button>
        <el-radio-button value="semantic">语义</el-radio-button>
      </el-radio-group>
    </div>
    <div
      class="tree-body"
      :class="{ 'drag-over-root': isDragging }"
      v-loading="loading ?? false"
      @dragover.prevent
      @drop="handleRootDrop"
    >
      <PageTreeNode
        v-for="node in tree"
        :key="node.id"
        :node="node"
        :selected-id="selectedId"
        :depth="0"
        @select="$emit('select', $event)"
        @create="$emit('create', $event)"
        @rename="$emit('rename', $event)"
        @delete="$emit('delete', $event)"
        @move="handleMove"
      />
      <el-empty v-if="!loading && tree.length === 0" description="暂无页面" :image-size="60">
        <el-button size="small" type="primary" @click="$emit('create', null)">新建页面</el-button>
      </el-empty>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, provide, ref, toRef } from 'vue'
import { Plus, Search } from '@element-plus/icons-vue'
import type { PageTreeNode as PageTreeNodeType } from '@/types/page'
import PageTreeNode from './PageTreeNode.vue'
import type { DropPosition } from '@/utils/tree'

const props = defineProps<{
  tree: PageTreeNodeType[]
  selectedId?: number | null
  loading?: boolean
  searchMode?: string
}>()

const emit = defineEmits<{
  select: [pageId: number]
  create: [parentId: number | null]
  rename: [node: PageTreeNodeType]
  delete: [node: PageTreeNodeType]
  move: [payload: { pageId: number; parentId: number | null; position: number }]
  search: [query: string]
  'update:searchMode': [mode: string]
}>()

const localSearchMode = computed({
  get: () => props.searchMode ?? 'keyword',
  set: (val: string) => emit('update:searchMode', val),
})

const searchText = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const draggedId = ref<number | null>(null)
const targetId = ref<number | null>(null)
const dropPosition = ref<DropPosition | null>(null)
const treeRef = toRef(props, 'tree')

const isDragging = computed(() => draggedId.value !== null)

provide('dragState', { draggedId, targetId, dropPosition, tree: treeRef })

function handleSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emit('search', searchText.value)
  }, 300)
}

function handleSearchClear() {
  if (debounceTimer) clearTimeout(debounceTimer)
  searchText.value = ''
  emit('search', '')
}

function handleMove(payload: { pageId: number; parentId: number | null; position: number }) {
  draggedId.value = null
  targetId.value = null
  dropPosition.value = null
  emit('move', payload)
}

function handleRootDrop(e: DragEvent) {
  e.preventDefault()
  if (draggedId.value === null) return
  emit('move', {
    pageId: draggedId.value,
    parentId: null,
    position: 9999,
  })
  draggedId.value = null
  targetId.value = null
  dropPosition.value = null
}
</script>

<style scoped>
.page-tree { height: 100%; display: flex; flex-direction: column; }
.tree-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; border-bottom: 1px solid var(--el-border-color-lighter); }
.tree-title { font-weight: 600; font-size: 13px; color: var(--el-text-color-secondary); }
.tree-search { padding: 4px 8px; border-bottom: 1px solid var(--el-border-color-lighter); }
.tree-body { flex: 1; overflow-y: auto; padding: 4px 0; }
.drag-over-root { background-color: var(--el-color-primary-light-9); }
</style>
