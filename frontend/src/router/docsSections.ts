// NanaFox fork: /docs 各小节的地址、标题和描述，路由、页面标题和文档页共用这一份。
// 每个小节一个独立地址（/docs/clients），方便搜索引擎分别收录；快速开始用 /docs 本身。
import type { RouteLocationNormalized } from 'vue-router'

export type DocsSection = 'quick-start' | 'claude-code' | 'codex-cli' | 'clients' | 'opencode' | 'api-usage' | 'faq'

export const DOCS_SECTIONS: DocsSection[] = ['quick-start', 'claude-code', 'codex-cli', 'clients', 'opencode', 'api-usage', 'faq']
export const DEFAULT_DOCS_SECTION: DocsSection = 'quick-start'
export const DOCS_ROUTE_NAME = 'Docs'

export const DOCS_SECTION_ROUTE_PATH = `/docs/:section(${DOCS_SECTIONS.filter(s => s !== DEFAULT_DOCS_SECTION).join('|')})?`

export function docsSectionPath(section: DocsSection): string {
  return section === DEFAULT_DOCS_SECTION ? '/docs' : `/docs/${section}`
}

export function toDocsSection(value: unknown): DocsSection | null {
  const cleaned = String(value ?? '').replace(/^#/, '').toLowerCase().trim()
  return DOCS_SECTIONS.includes(cleaned as DocsSection) ? (cleaned as DocsSection) : null
}

// 旧链接兼容：/docs#faq → /docs/faq（地址栏不留 #）
export function redirectLegacyDocsHash(to: RouteLocationNormalized) {
  if (to.params.section) return true
  const section = toDocsSection(to.hash)
  if (!section) return true
  return { path: docsSectionPath(section), query: to.query, hash: '', replace: true }
}

// 页面标题用各小节的大标题；/docs 本身沿用路由上的「文档」
const SECTION_TITLE_KEYS: Record<DocsSection, string> = {
  'quick-start': 'docs.title',
  'claude-code': 'docs.claudeCode.title',
  'codex-cli': 'docs.codex.title',
  'clients': 'docs.clients.title',
  'opencode': 'docs.opencode.title',
  'api-usage': 'docs.api.title',
  'faq': 'docs.faq.title'
}

export function resolveDocsSectionMetaKeys(section: unknown): { titleKey: string; descriptionKey: string } {
  const resolved = toDocsSection(section) ?? DEFAULT_DOCS_SECTION
  const camel = resolved.replace(/-([a-z])/g, (_, c: string) => c.toUpperCase())
  return { titleKey: SECTION_TITLE_KEYS[resolved], descriptionKey: `docs.seo.${camel}` }
}
