export interface Space {
  id: number
  name: string
  description: string
  icon: string
  cover: string
  visibility: 'private' | 'public'
  owner_id: number
  created_at: string
  updated_at: string
}

export interface SpaceMember {
  id: number
  space_id: number
  user_id: number
  role: 'admin' | 'member' | 'viewer'
  created_at: string
}

export interface CreateSpaceRequest {
  name: string
  description?: string
  icon?: string
  cover?: string
  visibility?: string
}
