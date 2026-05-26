import axios from 'axios'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

export interface DatabaseInput {
  host: string
  port: number
  username: string
  password: string
  dbname: string
}

export interface AdminInput {
  email: string
  username: string
  password: string
}

export interface MCPInput {
  enabled: boolean
  transport: string
  http_port: number
  api_keys: string[]
}

export interface RAGInput {
  enabled: boolean
  base_url: string
  api_key: string
}

export interface InstallRequest {
  database: DatabaseInput
  admin: AdminInput
  mcp?: MCPInput
  rag?: RAGInput
}

export interface InstallStatus {
  installed: boolean
}

export interface TestDBResult {
  connected: boolean
  version?: string
  error?: string
}

async function request<T>(method: string, url: string, data?: unknown): Promise<T> {
  const response = await axios.request<{ code: number; message: string; data: T }>({
    method,
    url: `${baseURL}${url}`,
    data,
    timeout: 30000,
  })
  return response.data.data
}

export function getInstallStatus(): Promise<InstallStatus> {
  return request<InstallStatus>('GET', '/install/status')
}

export function testDatabase(data: DatabaseInput): Promise<TestDBResult> {
  return request<TestDBResult>('POST', '/install/test-db', data)
}

export function executeInstall(data: InstallRequest): Promise<{ installed: boolean; message: string }> {
  return request<{ installed: boolean; message: string }>('POST', '/install/execute', data)
}
