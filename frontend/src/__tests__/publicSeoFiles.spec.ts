import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'
import { DOCS_SECTIONS, docsSectionPath, resolveDocsSectionMetaKeys } from '@/router/docsSections'
import zhDocs from '@/i18n/locales/zh/docs'
import enDocs from '@/i18n/locales/en/docs'

// NanaFox fork: robots.txt / sitemap.xml 放在 public/，打包后由后端当静态文件返回
const readPublic = (name: string) => readFileSync(resolve(__dirname, '../../public', name), 'utf8')

describe('public SEO files', () => {
  it('robots.txt allows the site and points to the sitemap', () => {
    const robots = readPublic('robots.txt')
    expect(robots).toMatch(/^User-agent: \*$/m)
    expect(robots).toMatch(/^Allow: \/$/m)
    expect(robots).toMatch(/^Disallow: \/api\/$/m)
    expect(robots).toContain('Sitemap: https://router.nanafox.com/sitemap.xml')
    expect(robots).not.toContain('<html')
  })

  it('sitemap.xml lists every docs section path', () => {
    const sitemap = readPublic('sitemap.xml')
    expect(sitemap.startsWith('<?xml')).toBe(true)
    for (const section of DOCS_SECTIONS) {
      expect(sitemap).toContain(`<loc>https://router.nanafox.com${docsSectionPath(section)}</loc>`)
    }
    expect(sitemap).not.toContain('#')
  })

  it('every docs section has a title and description in zh and en', () => {
    const lookup = (messages: Record<string, unknown>, key: string) =>
      key.replace(/^docs\./, '').split('.').reduce<unknown>((node, part) => (node as Record<string, unknown> | undefined)?.[part], messages)
    for (const section of DOCS_SECTIONS) {
      const { titleKey, descriptionKey } = resolveDocsSectionMetaKeys(section)
      for (const messages of [zhDocs, enDocs]) {
        expect(typeof lookup(messages, titleKey), `${section} ${titleKey}`).toBe('string')
        expect(typeof lookup(messages, descriptionKey), `${section} ${descriptionKey}`).toBe('string')
      }
    }
  })
})
