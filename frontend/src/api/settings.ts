import { api } from './client'

export interface MCPSettings {
  enabled: boolean
  transport: string
  http_port: number
  api_keys: string[]
  api_keys_masked: boolean
  default_user_id: number
}

export interface RAGSettings {
  enabled: boolean
  base_url: string
  api_key: string
  api_key_masked: boolean
  timeout: number
}

export interface SettingsResponse {
  mcp: MCPSettings
  rag: RAGSettings
}

export interface MCPSettingsRequest {
  enabled: boolean
  transport: string
  http_port: number
  api_keys: string[]
  api_keys_action: 'keep' | 'replace' | 'clear'
  default_user_id: number
}

export interface RAGSettingsRequest {
  enabled: boolean
  base_url: string
  api_key: string
  timeout: number
}

export interface SettingsRequest {
  mcp?: MCPSettingsRequest
  rag?: RAGSettingsRequest
}

export interface TestRAGResponse {
  connected: boolean
  message: string
}

export const UNCHANGED = '__UNCHANGED__'

export function getSettings(): Promise<SettingsResponse> {
  return api<SettingsResponse>('GET', '/admin/settings')
}

export function updateSettings(data: SettingsRequest): Promise<null> {
  return api<null>('PUT', '/admin/settings', data)
}

export function testRAGConnection(base_url: string, api_key: string, use_saved_key = false): Promise<TestRAGResponse> {
  return api<TestRAGResponse>('POST', '/admin/settings/test-rag', { base_url, api_key, use_saved_key })
}
