import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Page, PageTreeNode, UpdatePageRequest } from '@/types/page'
import * as pageApi from '@/api/page'

export const usePageStore = defineStore('page', () => {
  const currentPage = ref<Page | null>(null)
  const pageTree = ref<PageTreeNode[]>([])
  const loading = ref(false)

  async function fetchPage(spaceId: number, pageId: number) {
    loading.value = true
    try {
      currentPage.value = await pageApi.getPage(spaceId, pageId)
    } finally {
      loading.value = false
    }
  }

  async function fetchPageTree(spaceId: number) {
    loading.value = true
    try {
      pageTree.value = await pageApi.getPageTree(spaceId)
    } finally {
      loading.value = false
    }
  }

  async function createPage(spaceId: number, data: { title: string; content?: string; parent_id?: number | null }) {
    const page = await pageApi.createPage(spaceId, data)
    await fetchPageTree(spaceId)
    return page
  }

  async function updatePage(spaceId: number, pageId: number, data: UpdatePageRequest) {
    const page = await pageApi.updatePage(spaceId, pageId, data)
    if (currentPage.value?.id === pageId) {
      currentPage.value = page
    }
    await fetchPageTree(spaceId)
    return page
  }

  async function deletePage(spaceId: number, pageId: number) {
    await pageApi.deletePage(spaceId, pageId)
    if (currentPage.value?.id === pageId) {
      currentPage.value = null
    }
    await fetchPageTree(spaceId)
  }

  async function movePage(spaceId: number, pageId: number, data: { parent_id: number | null; position: number }) {
    const payload = { parent_id: data.parent_id ?? 0, position: data.position }
    const page = await pageApi.movePage(spaceId, pageId, payload)
    await fetchPageTree(spaceId)
    return page
  }

  return {
    currentPage, pageTree, loading,
    fetchPage, fetchPageTree, createPage, updatePage, deletePage, movePage,
  }
})
