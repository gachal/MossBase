<template>
  <div class="version-history" v-loading="loading">
    <el-page-header @back="router.push(`/spaces/${spaceId}/pages/${pageId}`)" title="返回页面" content="版本历史" />

    <div style="margin-top: 16px; display: flex; gap: 16px;">
      <div class="version-list" style="width: 300px;">
        <el-timeline>
          <el-timeline-item
            v-for="v in versions"
            :key="v.id"
            :timestamp="formatDate(v.created_at)"
            placement="top"
          >
            <el-card
              shadow="hover"
              :class="{ selected: selectedVersion === v.version_number }"
              @click="selectedVersion = v.version_number"
              style="cursor: pointer"
            >
              <div style="display: flex; justify-content: space-between; align-items: center;">
                <div>
                  <div style="font-weight: 600;">v{{ v.version_number }}</div>
                  <div style="font-size: 12px; color: var(--el-text-color-secondary);">{{ v.title }}</div>
                </div>
                <el-button
                  v-if="v.version_number !== currentVersion"
                  size="small"
                  type="primary"
                  text
                  @click.stop="handleRestore(v.version_number)"
                >还原</el-button>
              </div>
            </el-card>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-if="!loading && versions.length === 0" description="暂无版本记录" />
      </div>

      <div class="version-detail" style="flex: 1; min-width: 0;">
        <template v-if="diffResult">
          <h4>版本差异 (v{{ diffFrom }} → v{{ diffTo }})</h4>
          <div class="diff-output" v-html="diffHtml" />
        </template>
        <template v-else-if="selectedVersionData">
          <h4>v{{ selectedVersionData.version_number }} - {{ selectedVersionData.title }}</h4>
          <el-divider />
          <div class="version-content" v-html="renderContent(selectedVersionData.content)" />
        </template>
        <el-empty v-else description="选择一个版本查看详情" />
      </div>
    </div>

    <div style="margin-top: 16px;" v-if="versions.length >= 2">
      <el-button @click="showDiff" :disabled="!canDiff">对比选中版本</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api/client'

interface VersionItem {
  id: number
  page_id: number
  version_number: number
  title: string
  content: string
  edited_by: number
  created_at: string
}

const route = useRoute()
const router = useRouter()
const spaceId = computed(() => Number(route.params.id))
const pageId = computed(() => Number(route.params.pageId))

const loading = ref(false)
const versions = ref<VersionItem[]>([])
const selectedVersion = ref<number | null>(null)
const currentVersion = ref(0)
const diffResult = ref<string | null>(null)
const diffFrom = ref(0)
const diffTo = ref(0)

const selectedVersionData = computed(() => {
  if (!selectedVersion.value) return null
  return versions.value.find(v => v.version_number === selectedVersion.value) ?? null
})

const canDiff = computed(() => {
  return selectedVersion.value && versions.value.some(v => v.version_number !== selectedVersion.value)
})

onMounted(async () => {
  loading.value = true
  try {
    const data = await api<{ items: VersionItem[]; total: number }>('GET', `/spaces/${spaceId.value}/pages/${pageId.value}/versions`)
    versions.value = data.items
    if (versions.value.length > 0) {
      currentVersion.value = versions.value[0].version_number
      selectedVersion.value = versions.value[0].version_number
    }
  } catch {
    ElMessage.error('加载版本历史失败')
  } finally {
    loading.value = false
  }
})

async function showDiff() {
  if (!selectedVersion.value) return
  const other = versions.value.find(v => v.version_number !== selectedVersion.value)
  if (!other) return

  const from = Math.min(selectedVersion.value, other.version_number)
  const to = Math.max(selectedVersion.value, other.version_number)
  diffFrom.value = from
  diffTo.value = to

  try {
    const result = await api<{ from_version: number; to_version: number; diff: string }>(
      'GET', `/spaces/${spaceId.value}/pages/${pageId.value}/versions/diff?from=${from}&to=${to}`
    )
    diffResult.value = result.diff
  } catch {
    ElMessage.error('获取差异失败')
  }
}

const diffHtml = computed(() => {
  if (!diffResult.value) return ''
  try {
    const lines = JSON.parse(diffResult.value) as Array<{ Type: string; Text: string }>
    return lines.map(l => {
      const escaped = l.Text.replace(/</g, '&lt;').replace(/>/g, '&gt;')
      if (l.Type === 'added') return `<span class="diff-add">${escaped}</span>`
      if (l.Type === 'removed') return `<span class="diff-del">${escaped}</span>`
      return escaped
    }).join('')
  } catch {
    return diffResult.value
  }
})

async function handleRestore(versionNumber: number) {
  try {
    await ElMessageBox.confirm(`确定还原到 v${versionNumber}？将创建新版本。`, '还原版本', { type: 'warning' })
    await api('POST', `/spaces/${spaceId.value}/pages/${pageId.value}/versions/${versionNumber}/restore`)
    ElMessage.success('已还原')
    location.reload()
  } catch { /* cancelled */ }
}

function renderContent(content: string) {
  return content || '<em>无内容</em>'
}

function formatDate(d: string) {
  return new Date(d).toLocaleString('zh-CN')
}
</script>

<style scoped>
.version-history { padding: 24px; }
.selected { border-color: var(--el-color-primary) !important; }
.diff-output { font-family: monospace; white-space: pre-wrap; line-height: 1.6; padding: 12px; background: var(--el-fill-color-lighter); border-radius: 4px; }
:deep(.diff-add) { background: #d4edda; }
:deep(.diff-del) { background: #f8d7da; text-decoration: line-through; }
.version-content { line-height: 1.7; }
</style>
