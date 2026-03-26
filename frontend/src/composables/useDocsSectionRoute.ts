// frontend/src/composables/useDocsSectionRoute.ts
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'

export type DocsSection = 'quick-start' | 'claude-code' | 'codex-cli' | 'opencode' | 'api-usage' | 'faq'

const VALID_SECTIONS: DocsSection[] = ['quick-start', 'claude-code', 'codex-cli', 'opencode', 'api-usage', 'faq']
const DEFAULT_SECTION: DocsSection = 'quick-start'

function parseHash(hash: string): DocsSection {
  const cleaned = hash.replace(/^#/, '').toLowerCase().trim()
  if (VALID_SECTIONS.includes(cleaned as DocsSection)) return cleaned as DocsSection
  return DEFAULT_SECTION
}

const SECTION_TITLE_KEYS: Record<DocsSection, string> = {
  'quick-start': 'docs.nav.quickStart',
  'claude-code': 'docs.nav.claudeCode',
  'codex-cli': 'docs.nav.codexCli',
  'opencode': 'docs.nav.opencode',
  'api-usage': 'docs.nav.apiUsage',
  'faq': 'docs.nav.faq'
}

export function useDocsSectionRoute() {
  const route = useRoute()
  const router = useRouter()
  const { t } = useI18n()
  const appStore = useAppStore()

  const activeSection = computed<DocsSection>(() => parseHash(route.hash))

  function navigateTo(section: DocsSection) {
    router.replace({ hash: `#${section}` })
  }

  watch(activeSection, (section) => {
    const siteName = appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API'
    const sectionTitle = t(SECTION_TITLE_KEYS[section])
    const baseTitle = t('docs.title')
    document.title = `${sectionTitle} - ${baseTitle} - ${siteName}`
  }, { immediate: true })

  return {
    activeSection,
    navigateTo,
    VALID_SECTIONS
  }
}
