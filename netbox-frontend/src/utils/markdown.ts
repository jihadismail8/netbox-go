/**
 * Lightweight Markdown to HTML renderer.
 * Supports: headings, bold, italic, code, links, lists, blockquotes, hr.
 * No external dependencies. Escapes HTML for safety.
 */

function escapeHtml(text: string): string {
  const map: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  }
  return text.replace(/[&<>"']/g, (m) => map[m])
}

export function renderMarkdown(md: string): string {
  if (!md) return ''
  let text = md.replace(/\r\n/g, '\n')

  // Extract code blocks first (protect from other replacements)
  const codeBlocks: string[] = []
  text = text.replace(/```([\s\S]*?)```/g, (_m, code: string) => {
    const idx = codeBlocks.length
    codeBlocks.push(`<pre class="md-code"><code>${escapeHtml(code.trim())}</code></pre>`)
    return `\u0000CODEBLOCK${idx}\u0000`
  })

  // Escape HTML
  text = escapeHtml(text)

  // Headings
  text = text.replace(/^###### (.+)$/gm, '<h6>$1</h6>')
  text = text.replace(/^##### (.+)$/gm, '<h5>$1</h5>')
  text = text.replace(/^#### (.+)$/gm, '<h4>$1</h4>')
  text = text.replace(/^### (.+)$/gm, '<h3>$1</h3>')
  text = text.replace(/^## (.+)$/gm, '<h2>$1</h2>')
  text = text.replace(/^# (.+)$/gm, '<h1>$1</h1>')

  // Horizontal rule
  text = text.replace(/^---+$/gm, '<hr/>')

  // Blockquote
  text = text.replace(/^&gt; (.+)$/gm, '<blockquote>$1</blockquote>')

  // Bold
  text = text.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  // Italic
  text = text.replace(/\*(.+?)\*/g, '<em>$1</em>')
  // Inline code
  text = text.replace(/`([^`]+)`/g, '<code>$1</code>')
  // Links [text](url) — url was escaped, re-encode safely
  text = text.replace(
    /\[([^\]]+)\]\((https?:\/\/[^\s)]+)\)/g,
    '<a href="$2" target="_blank" rel="noopener noreferrer">$1</a>',
  )

  // Unordered lists
  text = text.replace(/(?:^|\n)((?:- .+(?:\n|$))+)/g, (_m, list: string) => {
    const items = list
      .trim()
      .split('\n')
      .map((l) => `<li>${l.replace(/^- /, '')}</li>`)
      .join('')
    return `\n<ul>${items}</ul>`
  })

  // Ordered lists
  text = text.replace(/(?:^|\n)((?:\d+\. .+(?:\n|$))+)/g, (_m, list: string) => {
    const items = list
      .trim()
      .split('\n')
      .map((l) => `<li>${l.replace(/^\d+\. /, '')}</li>`)
      .join('')
    return `\n<ol>${items}</ol>`
  })

  // Paragraphs — wrap lines that aren't already block elements
  const blocks = text.split(/\n\n+/)
  text = blocks
    .map((block) => {
      const trimmed = block.trim()
      if (!trimmed) return ''
      if (/^<(h[1-6]|ul|ol|pre|blockquote|hr|table|div)/.test(trimmed)) return trimmed
      if (trimmed.startsWith('\u0000CODEBLOCK')) return trimmed
      return `<p>${trimmed.replace(/\n/g, '<br/>')}</p>`
    })
    .join('\n')

  // Restore code blocks
  text = text.replace(
    /\u0000CODEBLOCK(\d+)\u0000/g,
    (_m, idx: string) => codeBlocks[Number(idx)] || '',
  )

  return text
}
