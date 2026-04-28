<template>
  <div class="page-editor" v-loading="pageStore.loading">
    <template v-if="pageStore.currentPage">
      <div class="editor-header">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="`/spaces/${spaceId}`">{{ spaceName }}</el-breadcrumb-item>
          <el-breadcrumb-item :to="`/spaces/${spaceId}/pages/${pageId}`">{{ pageStore.currentPage.title }}</el-breadcrumb-item>
        </el-breadcrumb>
        <div class="editor-actions">
          <span v-if="saving" class="save-status saving">保存中...</span>
          <span v-else-if="lastSaved" class="save-status saved">已保存</span>
          <el-button @click="handleSave" :loading="saving" type="primary">保存</el-button>
          <el-button @click="router.push(`/spaces/${spaceId}/pages/${pageId}`)">取消</el-button>
        </div>
      </div>

      <el-input
        v-model="title"
        class="title-input"
        placeholder="页面标题"
        size="large"
      />

      <MarkdownEditor v-model="content" @update:html="contentHtml = $event" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { usePageStore } from '@/stores/page'
import { useSpaceStore } from '@/stores/space'
import MarkdownEditor from '@/components/editor/MarkdownEditor.vue'

const route = useRoute()
const router = useRouter()
const pageStore = usePageStore()
const spaceStore = useSpaceStore()

const spaceId = computed(() => Number(route.params.id))
const pageId = computed(() => Number(route.params.pageId))
const spaceName = computed(() => spaceStore.currentSpace?.name ?? '空间')

const title = ref('')
const content = ref('')
const contentHtml = ref('')
const saving = ref(false)
const lastSaved = ref(false)

let autoSaveTimer: ReturnType<typeof setInterval> | null = null
let hasChanges = false
let initialized = false

onMounted(async () => {
  await pageStore.fetchPage(spaceId.value, pageId.value)
  if (!spaceStore.currentSpace || spaceStore.currentSpace.id !== spaceId.value) {
    await spaceStore.fetchSpace(spaceId.value)
  }
  if (pageStore.currentPage) {
    title.value = pageStore.currentPage.title
    content.value = pageStore.currentPage.content || ''
  }
  initialized = true

  autoSaveTimer = setInterval(() => {
    if (hasChanges) {
      handleAutoSave()
    }
  }, 30000)
})

onUnmounted(() => {
  if (autoSaveTimer) clearInterval(autoSaveTimer)
})

watch([title, content], () => {
  if (!initialized) return
  hasChanges = true
  lastSaved.value = false
})

async function handleAutoSave() {
  if (saving.value) return
  saving.value = true
  try {
    await pageStore.updatePage(spaceId.value, pageId.value, {
      title: title.value,
      content: content.value,
      content_html: contentHtml.value,
    })
    lastSaved.value = true
    hasChanges = false
  } catch {
    // silent fail for auto-save
  } finally {
    saving.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await pageStore.updatePage(spaceId.value, pageId.value, {
      title: title.value,
      content: content.value,
      content_html: contentHtml.value,
    })
    lastSaved.value = true
    hasChanges = false
    ElMessage.success('保存成功')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '保存失败'
    ElMessage.error(msg)
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.page-editor { padding: 24px; max-width: 900px; }
.editor-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.editor-actions { display: flex; align-items: center; gap: 8px; }
.save-status { font-size: 12px; }
.save-status.saving { color: var(--el-color-warning); }
.save-status.saved { color: var(--el-color-success); }
.title-input { margin-bottom: 16px; }
.title-input :deep(.el-input__inner) { font-size: 24px; font-weight: 600; }
</style>
