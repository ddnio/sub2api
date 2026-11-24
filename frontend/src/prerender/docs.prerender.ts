// NanaFox fork: 构建后给 /docs 各小节生成静态 HTML，不跑脚本的爬虫也能读到正文。
// 用线上同一套路由和页面组件渲染，内容只有一份来源。站点地址、接口地址、站点名、联系方式
// 这几个取自线上设置的值先留占位符，由后端按当时的设置填入（backend/internal/web/prerender.go）。
// 由 `pnpm run prerender`（vitest.prerender.config.ts）在 vite build 之后执行。
import { mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { expect, it } from 'vitest'
import { createApp, h, nextTick } from 'vue'
import { createPinia } from 'pinia'
import { RouterView } from 'vue-router'
import { DOCS_SECTIONS, docsSectionPath } from '@/router/docsSections'
import { PRERENDER_DIR, PRERENDER_MANIFEST, PRERENDER_PLACEHOLDERS, PRERENDER_SITE_ORIGIN } from './placeholders'

const distDir = resolve(__dirname, '../../../backend/internal/web/dist')
const APP_MOUNT = '<div id="app"></div>'

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, ch => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[ch] || ch)
}

it('prerenders every docs section', async () => {
  expect(window.location.origin).toBe(PRERENDER_SITE_ORIGIN)
  window.scrollTo = () => {} // jsdom 没实现，路由切换时会调用
  const indexHtml = readFileSync(resolve(distDir, 'index.html'), 'utf8')
  expect(indexHtml).toContain(APP_MOUNT)
  expect(indexHtml).toMatch(/<title>[^<]*<\/title>/)
  expect(indexHtml).not.toMatch(/<meta name="description"|rel="canonical"/)

  // 和 main.ts 一样：先装 pinia、读注入的设置，再装路由和多语言
  window.__APP_CONFIG__ = {
    site_name: PRERENDER_PLACEHOLDERS.siteName,
    api_base_url: PRERENDER_PLACEHOLDERS.apiBase,
    contact_info: PRERENDER_PLACEHOLDERS.contact,
  } as typeof window.__APP_CONFIG__
  const app = createApp({ render: () => h(RouterView) })
  app.use(createPinia())
  const { useAppStore } = await import('@/stores/app')
  useAppStore().initFromInjectedConfig()

  const { default: i18n, loadLocaleMessages } = await import('@/i18n')
  await loadLocaleMessages('zh')
  i18n.global.locale.value = 'zh'
  const { default: router } = await import('@/router')
  app.use(router)
  app.use(i18n)

  const root = document.createElement('div')
  document.body.appendChild(root)

  const pages: Record<string, string> = {}
  rmSync(resolve(distDir, PRERENDER_DIR), { recursive: true, force: true })
  mkdirSync(resolve(distDir, PRERENDER_DIR), { recursive: true })

  for (const section of DOCS_SECTIONS) {
    const path = docsSectionPath(section)
    await router.push(path)
    await router.isReady()
    if (!root.hasChildNodes()) app.mount(root)
    await nextTick()
    await new Promise(r => setTimeout(r, 0))
    expect(router.currentRoute.value.path).toBe(path)
    expect(root.textContent?.length ?? 0, `${path} has no content`).toBeGreaterThan(500)
    expect(document.head.querySelector('meta[name="description"]')?.getAttribute('content')).toBeTruthy()

    const head = Array.from(document.head.querySelectorAll('meta[name="description"], link[rel="canonical"]'))
      .map(el => el.outerHTML)
      .join('')
    expect(head).toContain(`${PRERENDER_SITE_ORIGIN}${path}"`)

    // 用回调替换：正文里的 $' $& 之类不能被当成替换语法
    const html = indexHtml
      .replace(/<title>[^<]*<\/title>/, () => `<title>${escapeHtml(document.title)}</title>${head}`)
      .replace(APP_MOUNT, () => `<div id="app">${root.innerHTML}</div>`)
    // 占位符只能原样出现；被转码或拼接后后端就替换不到了
    const stripped = Object.values(PRERENDER_PLACEHOLDERS).reduce((s, p) => s.split(p).join(''), html)
    expect(stripped, `${path} has transformed placeholders`).not.toMatch(/prerender\.invalid|__PRERENDER_/)
    expect(html).toContain(PRERENDER_PLACEHOLDERS.siteName)

    const file = `${section}.html`
    writeFileSync(resolve(distDir, PRERENDER_DIR, file), html)
    pages[path] = file
  }

  writeFileSync(resolve(distDir, PRERENDER_DIR, PRERENDER_MANIFEST), JSON.stringify({ placeholders: PRERENDER_PLACEHOLDERS, pages }, null, 2))
  app.unmount()
})
