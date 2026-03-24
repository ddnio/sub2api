import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import HomeView from '@/views/HomeView.vue'
import i18n, { loadLocaleMessages } from '@/i18n'
import { useAppStore } from '@/stores/app'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [
    { path: '/home', component: HomeView },
    { path: '/login', component: { template: '<div>login</div>' } },
    { path: '/dashboard', component: { template: '<div>dashboard</div>' } },
    { path: '/admin/dashboard', component: { template: '<div>admin-dashboard</div>' } },
    { path: '/docs', component: { template: '<div>docs</div>' } },
    { path: '/key-usage', component: { template: '<div>key-usage</div>' } }
  ]
})

describe('HomeView', () => {
  beforeEach(async () => {
    setActivePinia(createPinia())
    await loadLocaleMessages('zh')
    i18n.global.locale.value = 'zh'
    vi.restoreAllMocks()
    localStorage.clear()
  })

  async function mountView() {
    await router.push('/home')
    await router.isReady()

    const wrapper = mount(HomeView, {
      global: {
        plugins: [router, i18n],
        stubs: {
          LocaleSwitcher: true
        }
      }
    })

    await flushPromises()
    return wrapper
  }

  function seedDefaultSettings() {
    const appStore = useAppStore()
    appStore.publicSettingsLoaded = true
    appStore.siteName = 'Sub2API'
    appStore.apiBaseUrl = 'https://relay.sub2api.test'
    appStore.docUrl = ''
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
  }

  it('renders the default landing page with a public key-usage nav link', async () => {
    seedDefaultSettings()

    const wrapper = await mountView()
    const text = wrapper.text()

    expect(wrapper.find('a[href="/key-usage"]').exists()).toBe(true)
    expect(text).toContain('Sub2API')
    expect(text).toContain('统一接入你的 AI 上游')
    expect(text).toContain('立即开始')
    expect(text).toContain('查看文档')
    expect(text).toContain('替换基础 URL 即可接入')
    expect(text).toContain('OPENAI_BASE_URL')
  })

  it('renders iframe override when home_content is a URL', async () => {
    seedDefaultSettings()
    const appStore = useAppStore()
    appStore.cachedPublicSettings = {
      ...appStore.cachedPublicSettings!,
      home_content: 'https://example.com/embed'
    }

    const wrapper = await mountView()

    expect(wrapper.find('iframe').attributes('src')).toBe('https://example.com/embed')
  })

  it('renders HTML override when home_content is raw markup', async () => {
    seedDefaultSettings()
    const appStore = useAppStore()
    appStore.cachedPublicSettings = {
      ...appStore.cachedPublicSettings!,
      home_content: '<section id="custom-home">custom override</section>'
    }

    const wrapper = await mountView()

    expect(wrapper.html()).toContain('id="custom-home"')
    expect(wrapper.text()).toContain('custom override')
  })

  it('prefers external doc_url links over the internal docs route', async () => {
    seedDefaultSettings()
    const appStore = useAppStore()
    appStore.docUrl = 'https://docs.example.com'
    appStore.cachedPublicSettings = {
      ...appStore.cachedPublicSettings!,
      doc_url: 'https://docs.example.com'
    }

    const wrapper = await mountView()

    const externalDocLinks = wrapper
      .findAll('a')
      .filter((link) => link.attributes('href') === 'https://docs.example.com')

    expect(externalDocLinks.length).toBeGreaterThan(0)
  })

  it('uses the localized subtitle when public settings still contain the generic english default', async () => {
    seedDefaultSettings()
    const appStore = useAppStore()
    appStore.cachedPublicSettings = {
      ...appStore.cachedPublicSettings!,
      site_subtitle: 'Subscription to API Conversion Platform'
    }

    const wrapper = await mountView()

    expect(wrapper.text()).toContain('专为开发者打造的 API 中转服务')
    expect(wrapper.text()).not.toContain('Subscription to API Conversion Platform')
  })
})
