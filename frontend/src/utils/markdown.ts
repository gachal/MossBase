import { marked } from 'marked'
import hljs from 'highlight.js'

marked.use({
  renderer: {
    code({ text, lang }: { text: string; lang?: string }) {
      const language = lang && hljs.getLanguage(lang) ? lang : 'plaintext'
      const highlighted = hljs.highlight(text, { language }).value
      return `<pre><code class="hljs language-${language}">${highlighted}</code></pre>`
    },
  },
})

function preprocessMarkdown(raw: string): string {
  let r = raw
  r = r.replace(/\s*```(\w*)/g, '\n\n```$1\n')
  r = r.replace(/```\s*/g, '```\n\n')
  r = r.replace(/\s+\*\*/g, '\n\n**')
  r = r.replace(/(\*\*[^*]+\*\*)\s*(\|)/g, '$1\n\n$2')
  r = r.replace(/(\*\*[^*]+\*\*)\s+(- )/g, '$1\n$2')
  r = r.replace(/\| +\|/g, '|\n|')
  r = r.replace(/\n{3,}/g, '\n\n')
  return r
}

export function renderMarkdown(markdown: string): string {
  if (!markdown) return ''
  return marked.parse(preprocessMarkdown(markdown)) as string
}
