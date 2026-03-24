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
    expect(wrapper.text()).toContain('Codex CLI 使用指南')
    expect(wrapper.text()).toContain('安装 Node.js')
    expect(wrapper.text()).toContain('https://relay.sub2api.test/v1')
    expect(wrapper.text()).toContain('curl -I https://relay.sub2api.test/v1')
    expect(wrapper.text()).toContain('config.toml')
  })

  it('supports direct deep links while still rendering the same guide', async () => {
    const { wrapper } = await createWrapper('/docs?mode=codex')

    expect(wrapper.text()).toContain('codex --version')
    expect(wrapper.text()).toContain('auth.json')
    expect(wrapper.text()).toContain('常见问题')
    expect(wrapper.text()).toContain('model_provider = "sub2api"')
    expect(wrapper.text()).toContain('https://relay.sub2api.test/v1')
  })

  it('keeps rendering the same guide for legacy api mode links', async () => {
    const { wrapper, router } = await createWrapper('/docs?mode=api')

    expect(router.currentRoute.value.query.mode).toBe('api')
    expect(wrapper.text()).toContain('Codex CLI 使用指南')
  })
})
