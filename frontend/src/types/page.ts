export interface Page {
  id: number
  space_id: number
  parent_id: number | null
  title: string
  slug: string
  content: string
  content_html: string
  position: number
  status: 'draft' | 'published'
  version: number
  created_by: number
  updated_by: number
  created_at: string
  updated_at: string
}

export interface PageTreeNode extends Page {
  children: PageTreeNode[]
}

export interface PageVersion {
  id: number
  page_id: number
  version_number: number
  title: string
  content: string
  content_html: string
  edited_by: number
  created_at: string
}

export interface CreatePageRequest {
  space_id: number
  parent_id?: number | null
  title: string
  content?: string
}

export interface UpdatePageRequest {
  title?: string
  content?: string
  content_html?: string
  status?: string
}

export interface MovePageRequest {
  parent_id: number | null
  position: number
}
