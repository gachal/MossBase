import { api } from './client'
import type { PaginatedData } from '@/types/api'

export interface DashboardStats {
  total_users: number
  total_spaces: number
  total_pages: number
}

export interface AdminUser {
  id: number
  email: string
  username: string
  avatar: string
  role: string
  status: string
  created_at: string
}

export interface AdminSpace {
  id: number
  name: string
  description: string
  visibility: string
  owner_id: number
  created_at: string
  updated_at: string
  member_count?: number
  page_count?: number
}

export interface AdminPage {
  id: number
  space_id: number
  title: string
  slug: string
  status: string
  version: number
  created_by: number
  updated_by: number
  created_at: string
  updated_at: string
}

export function getDashboardStats() {
  return api<DashboardStats>('GET', '/admin/dashboard')
}

export function listAdminUsers(page = 1, pageSize = 20) {
  return api<PaginatedData<AdminUser>>('GET', `/admin/users?page=${page}&page_size=${pageSize}`)
}

export function updateUserRole(userId: number, role: string) {
  return api<null>('PUT', `/admin/users/${userId}/role`, { role })
}

export function updateUserStatus(userId: number, status: string) {
  return api<null>('PUT', `/admin/users/${userId}/status`, { status })
}

export function listAdminSpaces(page = 1, pageSize = 20) {
  return api<PaginatedData<AdminSpace>>('GET', `/admin/spaces?page=${page}&page_size=${pageSize}`)
}

export function getAdminSpaceDetail(spaceId: number) {
  return api<AdminSpace>('GET', `/admin/spaces/${spaceId}`)
}

export function deleteAdminSpace(spaceId: number) {
  return api<null>('DELETE', `/admin/spaces/${spaceId}`)
}

export function listAdminPages(page = 1, pageSize = 20) {
  return api<PaginatedData<AdminPage>>('GET', `/admin/pages?page=${page}&page_size=${pageSize}`)
}

export function deleteAdminPage(pageId: number) {
  return api<null>('DELETE', `/admin/pages/${pageId}`)
}
