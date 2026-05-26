import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Space, SpaceMember } from '@/types/space'
import { listSpaces, getSpace, listMembers } from '@/api/space'

export const useSpaceStore = defineStore('space', () => {
  const spaces = ref<Space[]>([])
  const total = ref(0)
  const currentSpace = ref<Space | null>(null)
  const members = ref<SpaceMember[]>([])
  const loading = ref(false)

  async function fetchSpaces(page = 1, pageSize = 20) {
    loading.value = true
    try {
      const result = await listSpaces(page, pageSize)
      spaces.value = result.items
      total.value = result.total
    } finally {
      loading.value = false
    }
  }

  async function fetchSpace(id: number) {
    loading.value = true
    try {
      currentSpace.value = await getSpace(id)
    } finally {
      loading.value = false
    }
  }

  async function fetchMembers(spaceId: number) {
    members.value = await listMembers(spaceId)
  }

  function userRoleInSpace(userId: number): string | null {
    const m = members.value.find(m => m.user_id === userId)
    return m?.role ?? null
  }

  function isAdmin(userId: number): boolean {
    return userRoleInSpace(userId) === 'admin'
  }

  return { spaces, total, currentSpace, members, loading, fetchSpaces, fetchSpace, fetchMembers, userRoleInSpace, isAdmin }
})
