export interface User {
  id: number
  email: string
  username: string
  avatar: string
  role: 'admin' | 'user'
  status: 'active' | 'disabled'
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}
