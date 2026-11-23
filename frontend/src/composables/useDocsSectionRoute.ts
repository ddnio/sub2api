// frontend/src/composables/useDocsSectionRoute.ts
import { computed, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  DEFAULT_DOCS_SECTION,
  DOCS_SECTIONS,
  docsSectionPath,
  resolveDocsSectionMetaKeys,
  toDocsSection,
  type DocsSection
} from '@/router/docsSections'

export type { DocsSection }
export { docsSectionPath }

function upsertHeadTag(selector: string, create: () => HTMLElement, attr: string, value: string) {
  let el = document.head.querySelector<HTMLElement>(selector)
  if (!el) {
    el = create()
    el.setAttribute('data-docs-seo', '')
    document.head.appendChild(el)
  }
  el.setAttribute(attr, value)
}

export function useDocsSectionRoute() {
  const route = useRoute()
  const router = useRouter()
  const { t, locale } = useI18n()

  const activeSection = computed<DocsSection>(() =>
    toDocsSection(route.params.section) ?? toDocsSection(route.hash) ?? DEFAULT_DOCS_SECTION
  )

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
    document.head.querySelectorAll('[data-docs-seo]').forEach(el => el.remove())
  })

  return {
    activeSection,
    navigateTo,
    VALID_SECTIONS: DOCS_SECTIONS
  }
}
