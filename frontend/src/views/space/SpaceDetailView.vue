<template>
  <div class="space-detail" v-loading="spaceStore.loading">
    <template v-if="spaceStore.currentSpace">
      <div class="space-header">
        <div>
          <h2>{{ spaceStore.currentSpace.name }}</h2>
        </div>
        <div>
          <el-button @click="router.push(`/spaces/${spaceId}/settings`)">空间设置</el-button>
        </div>
      </div>

      <div class="space-body">
        <div class="sidebar">
          <PageTree
            :tree="pageStore.pageTree"
            :selected-id="selectedPageId"
            :loading="pageStore.loading"
            v-model:search-mode="searchMode"
            @select="handleSelectPage"
            @create="handleCreateChild"
            @rename="handleRename"
            @delete="handleDeletePage"
            @move="handleMovePage"
            @search="handleSearch"
          />
        </div>
        <div class="main-content">
          <template v-if="searchQuery">
            <SpaceSearchResults
              :query="searchQuery"
              :mode="searchMode"
              :keyword-results="searchResults"
              :keyword-loading="searchLoading"
              :semantic-results="semanticResults"
              :semantic-loading="semanticLoading"
              @update:mode="searchMode = $event; handleSearch(searchQuery)"
              @select="handleSelectPage"
            />
          </template>
          <template v-else-if="selectedPageId">
            <router-view />
          </template>
          <el-empty v-else description="选择或创建一个页面开始" :image="emptyImage">
            <el-button type="primary" @click="handleCreateChild(null)">新建页面</el-button>
          </el-empty>
        </div>
      </div>

      <el-dialog v-model="showCreateDialog" :title="createParentId ? '新建子页面' : '新建页面'" width="480px">
        <el-form :model="createForm" label-width="80px">
          <el-form-item label="标题" required>
            <el-input v-model="createForm.title" placeholder="页面标题" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
        </template>
      </el-dialog>

      <el-dialog v-model="showRenameDialog" title="重命名页面" width="480px">
        <el-input v-model="renameTitle" placeholder="新标题" />
        <template #footer>
          <el-button @click="showRenameDialog = false">取消</el-button>
          <el-button type="primary" @click="submitRename">确定</el-button>
        </template>
      </el-dialog>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSpaceStore } from '@/stores/space'
import { usePageStore } from '@/stores/page'
import { searchPages, type SearchResultItem } from '@/api/page'
import { semanticSearch, type SemanticSearchResultItem } from '@/api/rag'
import type { PageTreeNode } from '@/types/page'
import PageTree from '@/components/page-tree/PageTree.vue'
import SpaceSearchResults from './SpaceSearchResults.vue'
import emptyImage from '@/assets/logo.png'

const route = useRoute()
const router = useRouter()
const spaceStore = useSpaceStore()
const pageStore = usePageStore()
const spaceId = computed(() => Number(route.params.id))
const selectedPageId = computed(() => {
  const p = route.params.pageId
  return p ? Number(p) : null
})

const showCreateDialog = ref(false)
const createParentId = ref<number | null>(null)
const creating = ref(false)
const createForm = reactive({ title: '' })

const showRenameDialog = ref(false)
const renameNode = ref<PageTreeNode | null>(null)
const renameTitle = ref('')

const searchQuery = ref('')
const searchResults = ref<SearchResultItem[]>([])
const searchLoading = ref(false)
const searchMode = ref<'keyword' | 'semantic'>('keyword')
const semanticResults = ref<SemanticSearchResultItem[]>([])
const semanticLoading = ref(false)

onMounted(async () => {
  await spaceStore.fetchSpace(spaceId.value)
  await pageStore.fetchPageTree(spaceId.value)
})

function handleSelectPage(pageId: number) {
  searchQuery.value = ''
  searchResults.value = []
  semanticResults.value = []
  router.push(`/spaces/${spaceId.value}/pages/${pageId}`)
}

async function handleSearch(query: string) {
  if (!query.trim()) {
    searchQuery.value = ''
    searchResults.value = []
    semanticResults.value = []
    return
  }
  searchQuery.value = query

  if (searchMode.value === 'keyword') {
    searchLoading.value = true
    try {
      const result = await searchPages(spaceId.value, query)
      searchResults.value = result.items ?? []
    } catch {
      searchResults.value = []
    } finally {
      searchLoading.value = false
    }
  } else {
    semanticLoading.value = true
    try {
      const result = await semanticSearch(spaceId.value, query)
      semanticResults.value = result.results ?? []
    } catch {
      semanticResults.value = []
    } finally {
      semanticLoading.value = false
    }
  }
}

function handleCreateChild(parentId: number | null) {
  createParentId.value = parentId
  createForm.title = ''
  showCreateDialog.value = true
}

async function submitCreate() {
  if (!createForm.title.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  creating.value = true
  try {
    const page = await pageStore.createPage(spaceId.value, {
      title: createForm.title,
      parent_id: createParentId.value,
    })
    showCreateDialog.value = false
    router.push(`/spaces/${spaceId.value}/pages/${page.id}/edit`)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '创建失败'
    ElMessage.error(msg)
  } finally {
    creating.value = false
  }
}

function handleRename(node: PageTreeNode) {
  renameNode.value = node
  renameTitle.value = node.title
  showRenameDialog.value = true
}

async function submitRename() {
  if (!renameNode.value || !renameTitle.value.trim()) return
  try {
    await pageStore.updatePage(spaceId.value, renameNode.value.id, { title: renameTitle.value })
    showRenameDialog.value = false
    ElMessage.success('重命名成功')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '重命名失败'
    ElMessage.error(msg)
  }
}

async function handleDeletePage(node: PageTreeNode) {
  try {
    await ElMessageBox.confirm(`确定删除「${node.title}」？`, '删除页面', { type: 'warning' })
    await pageStore.deletePage(spaceId.value, node.id)
    ElMessage.success('已删除')
    if (selectedPageId.value === node.id) {
      router.push(`/spaces/${spaceId.value}`)
    }
  } catch { /* cancelled */ }
}

async function handleMovePage({ pageId, parentId, position }: { pageId: number; parentId: number | null; position: number }) {
  try {
    await pageStore.movePage(spaceId.value, pageId, { parent_id: parentId, position })
    ElMessage.success('移动成功')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '移动失败'
    ElMessage.error(msg)
  }
}
</script>

<style scoped>
.space-detail { padding: 24px;padding-top: 0px; margin-top:-18px; }
.space-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 0px; }
.space-body { display: flex; gap: 16px; min-height: 500px; }
.sidebar { width: 260px; flex-shrink: 0; border: 1px solid var(--el-border-color-lighter); border-radius: 4px; overflow: hidden; }
.main-content { flex: 1; min-width: 0; }
</style>
