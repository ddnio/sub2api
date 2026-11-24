import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { h } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { RouterView, createMemoryHistory, createRouter } from 'vue-router'
import DocsView from '@/views/DocsView.vue'
import i18n, { loadLocaleMessages } from '@/i18n'
import { useAppStore } from '@/stores/app'
import { DOCS_SECTION_ROUTE_PATH, redirectLegacyDocsHash } from '@/router/docsSections'
import { buildOpenAICodexConfigToml } from '@/utils/codexAuthConfig'

const mounted: VueWrapper[] = []

async function createWrapper(startPath = '/docs') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/home', component: { template: '<div>home</div>' } },
      { path: DOCS_SECTION_ROUTE_PATH, name: 'Docs', component: DocsView, beforeEnter: redirectLegacyDocsHash },
      { path: '/batch-image', name: 'BatchImageGuide', alias: '/docs/batch-image', component: { template: '<div>batch</div>' } },
      { path: '/key-usage', component: { template: '<div>usage</div>' } },
      { path: '/login', component: { template: '<div>login</div>' } },
      { path: '/payment', component: { template: '<div>payment</div>' } }
    ]
  })

  await router.push(startPath)
  await router.isReady()

  // 通过 RouterView 挂载，和线上一样，页面内的路由钩子才会生效
  const wrapper = mount({ render: () => h(RouterView) }, {
    global: {
      plugins: [router, i18n],
      stubs: {
        LocaleSwitcher: true
      }
    }
  })

  await flushPromises()
  mounted.push(wrapper)

  return { wrapper, router }
}

describe('DocsView', () => {
  afterEach(() => {
    mounted.splice(0).forEach(w => w.unmount())
  })

  beforeEach(async () => {
    setActivePinia(createPinia())
    await loadLocaleMessages('zh')
    i18n.global.locale.value = 'zh'

    const appStore = useAppStore()
    appStore.publicSettingsLoaded = true
    appStore.siteName = 'Sub2API'
    appStore.apiBaseUrl = 'https://relay.sub2api.test'
    appStore.cachedPublicSettings = {
      site_name: 'Sub2API',
      site_logo: '',
      site_subtitle: '统一接入你的 AI 上游',
      api_base_url: 'https://relay.sub2api.test',
      contact_info: '',
      doc_url: '',
      home_content: '',
      registration_enabled: true,
      email_verify_enabled: false,
      registration_email_suffix_whitelist: [],
      promo_code_enabled: false,
      password_reset_enabled: false,
      invitation_code_enabled: false,
      turnstile_enabled: false,
      turnstile_site_key: '',
      hide_ccs_import_button: false,
      purchase_subscription_enabled: false,
      purchase_subscription_url: '',
      custom_menu_items: [],
      linuxdo_oauth_enabled: false,
      sora_client_enabled: false,
      backend_mode_enabled: false,
      version: 'test'
    }
  })

  it('renders the codex-style guide by default', async () => {
    const { wrapper } = await createWrapper()

    expect(wrapper.find('[data-docs-mode="api"]').exists()).toBe(false)
    expect(wrapper.find('[data-docs-mode="codex"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('快速开始')
    expect(wrapper.text()).toContain('三步接入')
    expect(wrapper.text()).toContain('https://relay.sub2api.test')
    expect(wrapper.text()).toContain('Codex CLI')

    const table = wrapper.find('[data-docs-group-table]')
    expect(table.text()).toContain('ClaudeCode专用')
    expect(table.text()).toContain('https://relay.sub2api.test/v1')
  })

  it('renders the desktop clients section with group-specific keys', async () => {
    const { wrapper } = await createWrapper('/docs#clients')

    expect(wrapper.find('[data-docs-client="cursor"]').exists()).toBe(true)
    expect(wrapper.find('[data-docs-client="cherry-studio"]').exists()).toBe(true)
    expect(wrapper.find('[data-docs-group-hint="anthropic"]').text()).toContain('不要加 /v1')
    expect(wrapper.find('[data-docs-group-hint="openai"]').text()).toContain('https://relay.sub2api.test/v1')
    expect(wrapper.find('[data-docs-client="cursor"]').text()).toContain('gpt-6.1-sol')

    const cherry = wrapper.find('[data-docs-client="cherry-studio"]')
    await cherry.findAll('button').find(b => b.text() === 'Claude')!.trigger('click')
    expect(cherry.find('pre').text()).toContain('claude-sonnet-5-5')
    expect(cherry.find('pre').text()).toMatch(/Base URL\s+https:\/\/relay\.sub2api\.test\n/)
  })

  it('does not tell codex users to append /v1 above the generated config', async () => {
    const { wrapper } = await createWrapper('/docs#codex-cli')

    const hint = wrapper.find('[data-docs-group-hint="openai"]').text()
    expect(hint).toContain('OpenAI GPT')
    expect(hint).not.toContain('/v1')
  })

  it('renders escaped @ in english copy', async () => {
    await loadLocaleMessages('en')
    i18n.global.locale.value = 'en'
    const { wrapper } = await createWrapper('/docs#opencode')

    expect(wrapper.text()).toContain('@ai-sdk/anthropic provider')
    expect(wrapper.text()).toContain('must include /v1')
  })

  it('keeps claude models out of the openai-compatible opencode config', async () => {
    const { wrapper } = await createWrapper('/docs#opencode')

    expect(wrapper.text()).toContain('gpt-6.1-sol')
    const configText = wrapper.findAll('pre').map(pre => pre.text()).join('\n')
    expect(configText).not.toContain('claude-sonnet')
  })

  it('shows both openai and anthropic api examples', async () => {
    const { wrapper } = await createWrapper('/docs#api-usage')

    expect(wrapper.text()).toContain('gpt-6.1-sol')
    expect(wrapper.text()).not.toContain('gpt-4o')
    const anthropic = wrapper.find('[data-docs-anthropic-example]')
    expect(anthropic.text()).toContain('claude-sonnet-5-5')
    expect(anthropic.text()).toContain('base_url="https://relay.sub2api.test"')
  })

  it('shows billing, error codes and contact in the faq', async () => {
    const { wrapper } = await createWrapper('/docs#faq')

    expect(wrapper.find('[data-docs-faq-billing]').text()).toContain('微信支付')
    expect(wrapper.find('[data-docs-faq-errors]').text()).toContain('401')
    expect(wrapper.find('[data-docs-faq-contact]').text()).toContain('nanafox-hi')
  })

  it('supports direct deep links while still rendering the same guide', async () => {
    const { wrapper } = await createWrapper('/docs#codex-cli')

    expect(wrapper.text()).toContain('codex --version')
    expect(wrapper.text()).toContain('config.toml')
    expect(wrapper.text()).toContain('常见问题')
    expect(wrapper.text()).toContain('model_provider = "OpenAI"')
    expect(wrapper.text()).toContain('experimental_bearer_token')
    expect(wrapper.text()).toContain('https://relay.sub2api.test/v1')
  })

  it('keeps rendering the same guide for legacy api mode links', async () => {
    const { wrapper, router } = await createWrapper('/docs?mode=api')

    expect(router.currentRoute.value.query.mode).toBe('api')
    expect(wrapper.text()).toContain('快速开始')
  })

  it('serves each section on its own path with canonical and description', async () => {
    const { wrapper, router } = await createWrapper('/docs/faq')

    expect(router.currentRoute.value.path).toBe('/docs/faq')
    expect(wrapper.find('[data-docs-faq-billing]').exists()).toBe(true)
    expect(document.head.querySelector('link[rel="canonical"]')?.getAttribute('href')).toBe(`${window.location.origin}/docs/faq`)
    expect(document.head.querySelector('meta[name="description"]')?.getAttribute('content')).toContain('401')
  })

  it('redirects legacy hash links to the section path', async () => {
    const { router } = await createWrapper('/docs#clients')

    expect(router.currentRoute.value.path).toBe('/docs/clients')
    expect(router.currentRoute.value.hash).toBe('')
  })

  it('renders crawlable nav links and navigates between sections', async () => {
    const { wrapper, router } = await createWrapper('/docs')

    const hrefs = wrapper.findAll('nav a').map(a => a.attributes('href'))
    expect(hrefs).toContain('/docs')
    expect(hrefs).toContain('/docs/clients')
    expect(document.head.querySelector('link[rel="canonical"]')?.getAttribute('href')).toBe(`${window.location.origin}/docs`)

    await wrapper.find('nav a[href="/docs/opencode"]').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.path).toBe('/docs/opencode')
    expect(wrapper.text()).toContain('opencode.json')
  })

  it('keeps /docs/batch-image on its own route and rejects unknown sections', async () => {
    const { router } = await createWrapper('/docs')

    expect(router.resolve('/docs/batch-image').name).toBe('BatchImageGuide')
    expect(router.resolve('/docs/unknown').matched).toHaveLength(0)
  })

  it('removes docs head tags when leaving the page', async () => {
    const { wrapper } = await createWrapper('/docs/faq')
    mounted.splice(mounted.indexOf(wrapper), 1)
    wrapper.unmount()

    expect(document.head.querySelector('link[rel="canonical"]')).toBeNull()
  })

  it('redirects legacy hash links while already on the docs page', async () => {
    const { router } = await createWrapper('/docs/clients')

    await router.push('/docs#faq')
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe('/docs/faq')

    await router.push('/docs#quick-start')
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe('/docs')
  })

  it('keeps the query string when following nav links', async () => {
    const { wrapper } = await createWrapper('/docs?utm_source=mail')

    expect(wrapper.find('nav a[href="/docs/faq?utm_source=mail"]').exists()).toBe(true)
  })

  it('updates the description when the locale changes', async () => {
    const { wrapper } = await createWrapper('/docs/faq')
    await loadLocaleMessages('en')
    i18n.global.locale.value = 'en'
    await wrapper.vm.$nextTick()

    expect(document.head.querySelector('meta[name="description"]')?.getAttribute('content')).toContain('Billing')
  })

  it('takes over prerendered head tags and removes them when leaving', async () => {
    const canonical = document.createElement('link')
    canonical.setAttribute('rel', 'canonical')
    canonical.setAttribute('data-docs-seo', '')
    canonical.setAttribute('href', 'https://router.example/docs/faq')
    document.head.appendChild(canonical)

    const { wrapper } = await createWrapper('/docs/faq')
    expect(document.head.querySelectorAll('link[rel="canonical"]')).toHaveLength(1)
    expect(canonical.getAttribute('href')).toBe(`${window.location.origin}/docs/faq`)
    mounted.splice(mounted.indexOf(wrapper), 1)
    wrapper.unmount()

    expect(document.head.querySelector('link[rel="canonical"]')).toBeNull()
  })

  it('restores a pre-existing description when leaving', async () => {
    const meta = document.createElement('meta')
    meta.setAttribute('name', 'description')
    meta.setAttribute('content', 'site description')
    document.head.appendChild(meta)

    const { wrapper } = await createWrapper('/docs/faq')
    expect(meta.getAttribute('content')).toContain('401')
    mounted.splice(mounted.indexOf(wrapper), 1)
    wrapper.unmount()

    expect(meta.getAttribute('content')).toBe('site description')
    expect(document.head.querySelectorAll('meta[name="description"]')).toHaveLength(1)
    meta.remove()
  })

  it('lists Codex CLI before Claude Code in nav and quick start, and shows quick jump card on codex docs', async () => {
    const { wrapper } = await createWrapper('/docs')
    const navHrefs = wrapper.findAll('nav a').map(a => a.attributes('href'))
    expect(navHrefs.indexOf('/docs/codex-cli')).toBeGreaterThan(-1)
    expect(navHrefs.indexOf('/docs/codex-cli')).toBeLessThan(navHrefs.indexOf('/docs/claude-code'))

    const quickStartLinks = wrapper.findAll('main a').map(a => a.attributes('href'))
    expect(quickStartLinks.indexOf('/docs/codex-cli')).toBeGreaterThan(-1)
    expect(quickStartLinks.indexOf('/docs/codex-cli')).toBeLessThan(quickStartLinks.indexOf('/docs/claude-code'))

    const { wrapper: codexWrapper } = await createWrapper('/docs/codex-cli')
    const card = codexWrapper.find('[data-testid="codex-quick-jump"]')
    expect(card.exists()).toBe(true)
    expect(card.text()).toContain('第 3 步')
    expect(card.text()).toContain('CC Switch')
    expect(card.html()).toContain('step=config')
    expect(card.html()).toContain('use=codex')
  })

  it('renders the Codex common errors FAQ using the shared codexAuthConfig generator', async () => {
    const { wrapper } = await createWrapper('/docs/faq')
    const codexFaq = wrapper.find('[data-docs-faq-codex-errors]')
    expect(codexFaq.exists()).toBe(true)
    expect(codexFaq.text()).toContain('Codex 常见报错')
    expect(codexFaq.text()).toContain('DAILY_LIMIT_EXCEEDED')
    expect(codexFaq.text()).toContain('No active subscription found for this group')
    expect(codexFaq.text()).toContain('previous_response_id')

    const expectedToml = buildOpenAICodexConfigToml({
      baseUrl: 'https://relay.sub2api.test',
      apiKey: 'your-api-key-here'
    })
    expect(codexFaq.find('pre').text()).toBe(expectedToml)
  })
})
