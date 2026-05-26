import axios from 'axios'
import type { ApiResponse } from '@/types/api'
import { getToken, removeToken } from '@/utils/storage'
import router from '@/router'

const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      removeToken()
      router.push('/login')
    }
    return Promise.reject(error)
  },
)

export async function api<T>(method: string, url: string, data?: unknown): Promise<T> {
  const response = await client.request<ApiResponse<T>>({ method, url, data, params: method === 'GET' ? data : undefined })
  return response.data.data
}

export default client
