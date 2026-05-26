import { marked } from 'marked'
import hljs from 'highlight.js'
import DOMPurify from 'dompurify'

export interface HeadingItem {
  text: string
  level: number
  id: string
}

export interface ProcessedMarkdown {
  html: string
  headings: HeadingItem[]
}

let headingOccurrences = new Map<string, number>()
let collectedHeadings: HeadingItem[] = []

function resetHeadingState() {
  headingOccurrences = new Map()
  collectedHeadings = []
}

function generateHeadingId(text: string): string {
  const base = text.toLowerCase().replace(/[^\w一-鿿]+/g, '-').replace(/^-|-$/g, '')
  const count = (headingOccurrences.get(base) ?? 0) + 1
  headingOccurrences.set(base, count)
  return count === 1 ? base : `${base}-${count}`
}

function escapeHtmlAttr(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
}

function stripMarkdownSyntax(text: string): string {
  return text
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/[*_`~#]/g, '')
    .trim()
}

function preprocessMarkdown(raw: string): string {
  let r = raw
  r = r.replace(/^\s*\[toc\]\s*$/gim, '')
  r = r.replace(/\s*```(\w*)/g, '\n\n```$1\n')
  r = r.replace(/```\s*/g, '```\n\n')
  r = r.replace(/(\*\*[^*]+\*\*)\s*(\|)/g, '$1\n\n$2')
  r = r.replace(/(\*\*[^*]+\*\*)\s+(- )/g, '$1\n$2')
  r = r.replace(/\| +\|/g, '|\n|')
  r = r.replace(/\n{3,}/g, '\n\n')
  return r
}

const SANITIZE_CONFIG = {
  ALLOWED_TAGS: [
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
    'p', 'br', 'hr',
    'ul', 'ol', 'li',
    'blockquote', 'pre', 'code',
    'strong', 'em', 'del', 's', 'mark',
    'a', 'img',
    'table', 'thead', 'tbody', 'tr', 'th', 'td',
    'sup', 'sub',
  ],
  ALLOWED_ATTR: ['href', 'src', 'alt', 'title', 'id', 'class', 'target', 'rel'],
}

marked.use({
  renderer: {
    heading({ text, depth }: { text: string; depth: number }) {
      const safeDepth = Math.max(1, Math.min(6, Math.floor(depth)))
      const id = generateHeadingId(text)
      collectedHeadings.push({ text: stripMarkdownSyntax(text), level: safeDepth, id })
      const escapedId = escapeHtmlAttr(id)
      const content = marked.parseInline(text) as string
      return `<h${safeDepth} id="${escapedId}">${content}</h${safeDepth}>`
    },
    code({ text, lang }: { text: string; lang?: string }) {
      const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
      const highlighted = hljs.highlight(text, { language }).value
      return `<pre><code class="hljs language-${language}">${highlighted}</code></pre>`
    },
  },
})

/** Single-pass: returns both rendered HTML and heading list with consistent IDs */
export function processMarkdown(markdown: string): ProcessedMarkdown {
  if (!markdown) return { html: '', headings: [] }
  resetHeadingState()
  const raw = marked.parse(preprocessMarkdown(markdown)) as string
  const html = DOMPurify.sanitize(raw, SANITIZE_CONFIG)
  return { html, headings: collectedHeadings.map(h => ({ ...h })) }
}

/** Render markdown to sanitized HTML (uses single-pass internally) */
export function renderMarkdown(markdown: string): string {
  return processMarkdown(markdown).html
}

/** Extract heading list from markdown (independent pass, IDs match processMarkdown) */
export function extractHeadings(markdown: string): HeadingItem[] {
  if (!markdown) return []
  resetHeadingState()
  const tokens = marked.lexer(preprocessMarkdown(markdown))
  const headings: HeadingItem[] = []
  for (const token of tokens) {
    if (token.type === 'heading') {
      const id = generateHeadingId(token.text)
      headings.push({ text: stripMarkdownSyntax(token.text), level: Math.max(1, Math.min(6, token.depth)), id })
    }
  }
  return headings
}
