import { beforeEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import DocsView from '@/views/DocsView.vue'
import i18n, { loadLocaleMessages } from '@/i18n'
import { useAppStore } from '@/stores/app'

async function createWrapper(startPath = '/docs') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/home', component: { template: '<div>home</div>' } },
      { path: '/docs', component: DocsView },
      { path: '/key-usage', component: { template: '<div>usage</div>' } },
      { path: '/login', component: { template: '<div>login</div>' } },
      { path: '/payment', component: { template: '<div>payment</div>' } }
    ]
  })

  await router.push(startPath)
  await router.isReady()

  const wrapper = mount(DocsView, {
    global: {
      plugins: [router, i18n],
      stubs: {
        LocaleSwitcher: true
      }
    }
  })

  await flushPromises()

  return { wrapper, router }
}

describe('DocsView', () => {
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
    expect(wrapper.text()).toContain('auth.json')
    expect(wrapper.text()).toContain('常见问题')
    expect(wrapper.text()).toContain('model_provider = "sub2api"')
    expect(wrapper.text()).toContain('https://relay.sub2api.test')
  })

  it('keeps rendering the same guide for legacy api mode links', async () => {
    const { wrapper, router } = await createWrapper('/docs?mode=api')

    expect(router.currentRoute.value.query.mode).toBe('api')
    expect(wrapper.text()).toContain('快速开始')
  })
})
