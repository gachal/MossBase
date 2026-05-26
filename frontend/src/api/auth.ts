import { api } from './client'
import type { LoginRequest, RegisterRequest, LoginResponse, User } from '@/types/user'

export function register(data: RegisterRequest) {
  return api<LoginResponse>('POST', '/auth/register', data)
}

export function login(data: LoginRequest) {
  return api<LoginResponse>('POST', '/auth/login', data)
}

export function getProfile() {
  return api<User>('GET', '/user/profile')
}

export function updateProfile(data: { username?: string; avatar?: string }) {
  return api<User>('PUT', '/user/profile', data)
}
