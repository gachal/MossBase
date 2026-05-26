<template>
  <Teleport to="body">
    <div class="context-overlay" @click="$emit('close')" @contextmenu.prevent="$emit('close')">
      <div class="context-menu" :style="{ left: x + 'px', top: y + 'px' }">
        <div class="menu-item" @click="$emit('create-child')">
          <el-icon><Plus /></el-icon> 新建子页面
        </div>
        <div class="menu-item" @click="$emit('rename')">
          <el-icon><Edit /></el-icon> 重命名
        </div>
        <div class="menu-divider" />
        <div class="menu-item danger" @click="$emit('delete')">
          <el-icon><Delete /></el-icon> 删除
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { Plus, Edit, Delete } from '@element-plus/icons-vue'
import type { PageTreeNode } from '@/types/page'

defineProps<{ x: number; y: number; node: PageTreeNode }>()
defineEmits<{
  close: []
  'create-child': []
  rename: []
  delete: []
}>()
</script>

<style scoped>
.context-overlay { position: fixed; inset: 0; z-index: 2000; }
.context-menu {
  position: fixed; z-index: 2001;
  background: var(--el-bg-color-overlay); border-radius: 4px;
  box-shadow: var(--el-box-shadow-light); padding: 4px 0; min-width: 140px;
}
.menu-item {
  display: flex; align-items: center; gap: 8px;
  padding: 6px 16px; cursor: pointer; font-size: 13px;
}
.menu-item:hover { background: var(--el-fill-color-light); }
.menu-item.danger { color: var(--el-color-danger); }
.menu-item.danger:hover { background: var(--el-color-danger-light-9); }
.menu-divider { height: 1px; margin: 4px 0; background: var(--el-border-color-lighter); }
</style>
