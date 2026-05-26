import { api } from './client'
import type { Page, PageTreeNode, CreatePageRequest, UpdatePageRequest, MovePageRequest } from '@/types/page'

export function createPage(spaceId: number, data: Omit<CreatePageRequest, 'space_id'>) {
  return api<Page>('POST', `/spaces/${spaceId}/pages`, data)
}

export function getPage(spaceId: number, pageId: number) {
  return api<Page>('GET', `/spaces/${spaceId}/pages/${pageId}`)
}

export function updatePage(spaceId: number, pageId: number, data: UpdatePageRequest) {
  return api<Page>('PUT', `/spaces/${spaceId}/pages/${pageId}`, data)
}

export function deletePage(spaceId: number, pageId: number) {
  return api<null>('DELETE', `/spaces/${spaceId}/pages/${pageId}`)
}

export function getPageTree(spaceId: number) {
  return api<PageTreeNode[]>('GET', `/spaces/${spaceId}/pages/tree`)
}

export function movePage(spaceId: number, pageId: number, data: MovePageRequest) {
  return api<Page>('PUT', `/spaces/${spaceId}/pages/${pageId}/move`, data)
}

export interface SearchResultItem {
  id: number
  title: string
  snippet: string
  updated_at: string
}

export interface SearchResult {
  items: SearchResultItem[]
  total: number
  page: number
  page_size: number
}

export function searchPages(spaceId: number, query: string) {
  return api<SearchResult>('GET', `/spaces/${spaceId}/pages/search?q=${encodeURIComponent(query)}`)
}
