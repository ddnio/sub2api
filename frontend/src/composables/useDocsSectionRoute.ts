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

// 自己新建的、以及预渲染页面带来的（带 data-docs-seo）标签离开时删掉；
// 页面上原本就有的只改值并记下原值，离开时还原
function upsertHeadTag(restoreHeadTags: Map<HTMLElement, () => void>, selector: string, create: () => HTMLElement, attr: string, value: string) {
  let el = document.head.querySelector<HTMLElement>(selector)
  const owned = !el || el.hasAttribute('data-docs-seo')
  if (!el) {
    el = create()
    document.head.appendChild(el)
  }
  if (!restoreHeadTags.has(el)) {
    const tag = el
    const original = tag.getAttribute(attr)
    restoreHeadTags.set(tag, owned ? () => tag.remove() : () => {
      if (original === null) tag.removeAttribute(attr)
      else tag.setAttribute(attr, original)
      tag.removeAttribute('data-docs-seo')
    })
  }
  el.setAttribute('data-docs-seo', '')
  el.setAttribute(attr, value)
}

export function useDocsSectionRoute() {
  const route = useRoute()
  const router = useRouter()
  const { t, locale } = useI18n()
  const restoreHeadTags = new Map<HTMLElement, () => void>()

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
    upsertHeadTag(restoreHeadTags, 'meta[name="description"]', () => {
      const m = document.createElement('meta')
      m.setAttribute('name', 'description')
      return m
    }, 'content', t(resolveDocsSectionMetaKeys(section).descriptionKey))
    upsertHeadTag(restoreHeadTags, 'link[rel="canonical"]', () => {
      const l = document.createElement('link')
      l.setAttribute('rel', 'canonical')
      return l
    }, 'href', `${window.location.origin}${docsSectionPath(section)}`)
  }, { immediate: true })

  // 离开文档页时去掉这里加的描述和规范地址，免得留给别的页面
  onBeforeUnmount(() => {
    restoreHeadTags.forEach(restore => restore())
    restoreHeadTags.clear()
  })

  return {
    activeSection,
    navigateTo,
    VALID_SECTIONS: DOCS_SECTIONS
  }
}
