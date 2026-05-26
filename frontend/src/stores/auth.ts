import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/user'
import { getToken, setToken, removeToken, getUser, setUser } from '@/utils/storage'
import { login as apiLogin, register as apiRegister, getProfile } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(getUser<User>())
  const token = ref<string | null>(getToken())

  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(email: string, password: string) {
    const result = await apiLogin({ email, password })
    user.value = result.user
    token.value = result.token
    setUser(result.user)
    setToken(result.token)
  }

  async function register(email: string, username: string, password: string) {
    const result = await apiRegister({ email, username, password })
    user.value = result.user
    token.value = result.token
    setUser(result.user)
    setToken(result.token)
  }

  async function fetchProfile() {
    const profile = await getProfile()
    user.value = profile
    setUser(profile)
  }

  function clearAuth() {
    user.value = null
    token.value = null
    removeToken()
  }

  return { user, token, isAuthenticated, isAdmin, login, register, fetchProfile, clearAuth }
})
