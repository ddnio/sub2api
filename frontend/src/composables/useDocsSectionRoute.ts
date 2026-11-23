// frontend/src/composables/useDocsSectionRoute.ts
import { computed, onBeforeUnmount, watch } from 'vue'
import { onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  DEFAULT_DOCS_SECTION,
  DOCS_SECTIONS,
  docsSectionPath,
  redirectLegacyDocsHash,
  resolveDocsSectionMetaKeys,
  toDocsSection,
  type DocsSection
} from '@/router/docsSections'

export type { DocsSection }
export { docsSectionPath }

// 页面上原本就有的标签只改值并记下原值，离开时还原；自己新建的离开时删掉
const restoreHeadTags: Array<() => void> = []

function upsertHeadTag(selector: string, create: () => HTMLElement, attr: string, value: string) {
  let el = document.head.querySelector<HTMLElement>(selector)
  if (!el) {
    el = create()
    document.head.appendChild(el)
    const created = el
    restoreHeadTags.push(() => created.remove())
  } else if (!el.hasAttribute('data-docs-seo')) {
    const existing = el
    const original = existing.getAttribute(attr)
    restoreHeadTags.push(() => {
      if (original === null) existing.removeAttribute(attr)
      else existing.setAttribute(attr, original)
      existing.removeAttribute('data-docs-seo')
    })
  }
  el.setAttribute('data-docs-seo', '')
  el.setAttribute(attr, value)
}

export function useDocsSectionRoute() {
  const route = useRoute()
  const router = useRouter()
  const { t, locale } = useI18n()

  const activeSection = computed<DocsSection>(() =>
    toDocsSection(route.params.section) ?? toDocsSection(route.hash) ?? DEFAULT_DOCS_SECTION
  )

  // 已经在文档页时点到旧的 #小节 链接，beforeEnter 不会触发，这里补上
  onBeforeRouteUpdate(redirectLegacyDocsHash)

  function navigateTo(section: DocsSection) {
    router.push({ path: docsSectionPath(section), query: route.query })
  }

  // 页面标题由 router/title.ts 统一生成；这里只补搜索引擎要看的描述和规范地址
  watch([activeSection, locale], ([section]) => {
    upsertHeadTag('meta[name="description"]', () => {
      const m = document.createElement('meta')
      m.setAttribute('name', 'description')
      return m
    }, 'content', t(resolveDocsSectionMetaKeys(section).descriptionKey))
    upsertHeadTag('link[rel="canonical"]', () => {
      const l = document.createElement('link')
      l.setAttribute('rel', 'canonical')
      return l
    }, 'href', `${window.location.origin}${docsSectionPath(section)}`)
  }, { immediate: true })

  // 离开文档页时去掉这里加的描述和规范地址，免得留给别的页面
  onBeforeUnmount(() => {
    restoreHeadTags.splice(0).reverse().forEach(restore => restore())
  })

  return {
    activeSection,
    navigateTo,
    VALID_SECTIONS: DOCS_SECTIONS
  }
}
