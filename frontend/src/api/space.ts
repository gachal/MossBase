import { api } from './client'
import type { Space, SpaceMember, CreateSpaceRequest } from '@/types/space'
import type { PaginatedData } from '@/types/api'

export function createSpace(data: CreateSpaceRequest) {
  return api<Space>('POST', '/spaces', data)
}

export function listSpaces(page = 1, pageSize = 20) {
  return api<PaginatedData<Space>>('GET', `/spaces?page=${page}&page_size=${pageSize}`)
}

export function getSpace(id: number) {
  return api<Space>('GET', `/spaces/${id}`)
}

export function updateSpace(id: number, data: Partial<CreateSpaceRequest>) {
  return api<Space>('PUT', `/spaces/${id}`, data)
}

export function deleteSpace(id: number) {
  return api<null>('DELETE', `/spaces/${id}`)
}

export function listMembers(spaceId: number) {
  return api<SpaceMember[]>('GET', `/spaces/${spaceId}/members`)
}

export function addMember(spaceId: number, userId: number, role: string) {
  return api<null>('POST', `/spaces/${spaceId}/members`, { user_id: userId, role })
}

export function removeMember(spaceId: number, userId: number) {
  return api<null>('DELETE', `/spaces/${spaceId}/members/${userId}`)
}
