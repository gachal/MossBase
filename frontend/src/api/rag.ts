import { api } from './client'

export interface SemanticSearchResultItem {
  document_id: string
  title: string
  content: string
  score: number
}

export interface SemanticSearchResponse {
  results: SemanticSearchResultItem[]
  total: number
  query: string
}

export function semanticSearch(spaceId: number, query: string, limit = 10): Promise<SemanticSearchResponse> {
  return api<SemanticSearchResponse>(
    'GET',
    `/spaces/${spaceId}/pages/semantic-search?q=${encodeURIComponent(query)}&limit=${limit}`
  )
}
