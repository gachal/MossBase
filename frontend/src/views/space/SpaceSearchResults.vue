<template>
  <div class="search-results">
    <div class="search-header">
      <span>搜索 "{{ query }}" 的结果</span>
      <div class="search-header-right">
        <el-radio-group :model-value="mode" size="small" @change="(v?: string | number | boolean) => { if (v) $emit('update:mode', v as 'keyword' | 'semantic') }">
          <el-radio-button value="keyword">关键词</el-radio-button>
          <el-radio-button value="semantic">语义</el-radio-button>
        </el-radio-group>
        <span class="search-count">
          {{ mode === 'keyword' ? keywordResults.length : semanticResults.length }} 条结果
        </span>
      </div>
    </div>

    <!-- Keyword search results -->
    <template v-if="mode === 'keyword'">
      <div v-loading="keywordLoading">
        <div v-if="keywordResults.length === 0 && !keywordLoading" class="search-empty">
          <el-empty description="未找到匹配的页面" :image-size="80" />
        </div>
        <div
          v-for="item in keywordResults"
          :key="item.id"
          class="search-item"
          @click="$emit('select', item.id)"
        >
          <div class="search-item-title">{{ item.title }}</div>
          <div class="search-item-snippet" v-html="highlightKeyword(item.snippet, query)"></div>
          <div class="search-item-meta">{{ formatDate(item.updated_at) }}</div>
        </div>
      </div>
    </template>

    <!-- Semantic search results -->
    <template v-else>
      <div v-loading="semanticLoading">
        <div v-if="semanticResults.length === 0 && !semanticLoading" class="search-empty">
          <el-empty description="未找到语义匹配的页面" :image-size="80" />
        </div>
        <div
          v-for="item in semanticResults"
          :key="item.document_id"
          class="search-item"
          @click="$emit('select', Number(item.document_id))"
        >
          <div class="search-item-title">{{ item.title }}</div>
          <div class="search-item-snippet">{{ truncateContent(item.content) }}</div>
          <div class="search-item-score">
            <span class="score-label">相关度</span>
            <el-progress
              :percentage="Math.round(item.score * 100)"
              :stroke-width="8"
              :color="getScoreColor(item.score)"
              style="width: 120px"
            />
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify'
import type { SearchResultItem } from '@/api/page'
import type { SemanticSearchResultItem } from '@/api/rag'

defineProps<{
  query: string
  mode: 'keyword' | 'semantic'
  keywordResults: SearchResultItem[]
  keywordLoading: boolean
  semanticResults: SemanticSearchResultItem[]
  semanticLoading: boolean
}>()

defineEmits<{
  'update:mode': [value: 'keyword' | 'semantic']
  'select': [pageId: number]
}>()

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
}

function highlightKeyword(text: string, keyword: string): string {
  if (!text || !keyword) return escapeHtml(text || '')
  const safeText = escapeHtml(text)
  const safeKeyword = escapeHtml(keyword)
  const escaped = safeKeyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const regex = new RegExp(`(${escaped})`, 'gi')
  const highlighted = safeText.replace(regex, '<mark>$1</mark>')
  return DOMPurify.sanitize(highlighted, { ALLOWED_TAGS: ['mark'] })
}

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function truncateContent(content: string, maxLen = 200): string {
  if (!content) return ''
  return content.length > maxLen ? content.slice(0, maxLen) + '...' : content
}

function getScoreColor(score: number): string {
  if (score >= 0.8) return '#67c23a'
  if (score >= 0.5) return '#e6a23c'
  return '#f56c6c'
}
</script>

<style scoped>
.search-results { padding: 16px; }
.search-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; font-size: 14px; color: var(--el-text-color-regular); }
.search-header-right { display: flex; align-items: center; gap: 12px; }
.search-count { color: var(--el-text-color-secondary); font-size: 13px; }
.search-empty { padding: 40px 0; }
.search-item { padding: 12px 16px; margin-bottom: 8px; border: 1px solid var(--el-border-color-lighter); border-radius: 6px; cursor: pointer; transition: border-color 0.2s; }
.search-item:hover { border-color: var(--el-color-primary); }
.search-item-title { font-size: 15px; font-weight: 600; color: var(--el-text-color-primary); margin-bottom: 6px; }
.search-item-snippet { font-size: 13px; color: var(--el-text-color-secondary); line-height: 1.5; margin-bottom: 6px; }
.search-item-snippet :deep(mark) { background-color: var(--el-color-warning-light-7); color: var(--el-text-color-primary); padding: 0 2px; border-radius: 2px; }
.search-item-meta { font-size: 12px; color: var(--el-text-color-placeholder); }
.search-item-score { display: flex; align-items: center; gap: 8px; margin-top: 4px; }
.score-label { font-size: 12px; color: var(--el-text-color-secondary); white-space: nowrap; }
</style>
