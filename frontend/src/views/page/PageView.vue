<template>
  <div class="page-view" v-loading="pageStore.loading">
    <template v-if="pageStore.currentPage">
      <div class="page-header">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="`/spaces/${spaceId}`">{{ spaceName }}</el-breadcrumb-item>
          <el-breadcrumb-item>{{ pageStore.currentPage.title }}</el-breadcrumb-item>
        </el-breadcrumb>
        <div class="page-actions">
          <el-button type="primary" @click="router.push(`/spaces/${spaceId}/pages/${pageId}/edit`)">编辑</el-button>
        </div>
      </div>
      <h1 class="page-title">{{ pageStore.currentPage.title }}</h1>
      <div class="page-meta">
        <span>版本 {{ pageStore.currentPage.version }}</span>
        <span>更新于 {{ formatDate(pageStore.currentPage.updated_at) }}</span>
        <router-link :to="`/spaces/${spaceId}/pages/${pageId}/versions`">历史版本</router-link>
      </div>
      <el-divider />
      <div class="page-content" v-html="renderedContent" />
    </template>
    <el-empty v-else-if="!pageStore.loading" description="页面不存在" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { usePageStore } from '@/stores/page'
import { useSpaceStore } from '@/stores/space'
import { renderMarkdown } from '@/utils/markdown'

const route = useRoute()
const router = useRouter()
const pageStore = usePageStore()
const spaceStore = useSpaceStore()

const spaceId = computed(() => Number(route.params.id))
const pageId = computed(() => Number(route.params.pageId))
const spaceName = computed(() => spaceStore.currentSpace?.name ?? '空间')

const renderedContent = computed(() => {
  return renderMarkdown(pageStore.currentPage?.content ?? '')
})

function loadPage() {
  if (pageId.value) {
    pageStore.fetchPage(spaceId.value, pageId.value)
  }
}

watch(pageId, () => {
  loadPage()
})

onMounted(() => {
  loadPage()
  if (!spaceStore.currentSpace || spaceStore.currentSpace.id !== spaceId.value) {
    spaceStore.fetchSpace(spaceId.value)
  }
})

function formatDate(d: string) {
  return new Date(d).toLocaleString('zh-CN')
}
</script>

<style scoped>
.page-view { padding: 24px; max-width: 900px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.page-title { margin: 0 0 8px; font-size: 28px; }
.page-meta { display: flex; gap: 16px; font-size: 13px; color: var(--el-text-color-secondary); }
.page-meta a { color: var(--el-color-primary); text-decoration: none; }
.page-content { line-height: 1.7; }
.page-content :deep(h1) { font-size: 24px; margin: 24px 0 12px; }
.page-content :deep(h2) { font-size: 20px; margin: 20px 0 10px; }
.page-content :deep(h3) { font-size: 17px; margin: 16px 0 8px; }
.page-content :deep(p) { margin: 8px 0; }
.page-content :deep(ul), .page-content :deep(ol) { padding-left: 24px; margin: 8px 0; }
.page-content :deep(blockquote) { border-left: 4px solid var(--el-border-color); padding-left: 16px; color: var(--el-text-color-secondary); margin: 8px 0; }
.page-content :deep(pre) { background: var(--el-fill-color-lighter); padding: 12px; border-radius: 4px; overflow-x: auto; }
.page-content :deep(code) { font-family: 'Menlo', 'Monaco', monospace; font-size: 13px; }
.page-content :deep(pre code) { background: none; padding: 0; }
.page-content :deep(code) { background: var(--el-fill-color); padding: 2px 4px; border-radius: 2px; }
.page-content :deep(table) { border-collapse: collapse; width: 100%; margin: 8px 0; }
.page-content :deep(td), .page-content :deep(th) { border: 1px solid var(--el-border-color); padding: 8px; min-width: 80px; }
.page-content :deep(th) { background: var(--el-fill-color-lighter); font-weight: 600; }
.page-content :deep(img) { max-width: 100%; height: auto; border-radius: 4px; }
.page-content :deep(hr) { border: none; border-top: 2px solid var(--el-border-color); margin: 16px 0; }
</style>
