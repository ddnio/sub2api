import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { ref } from 'vue'
import FloatingContactButton from '@/components/FloatingContactButton.vue'
import { useAppStore } from '@/stores/app'

const mockRoutePath = ref('/')

vi.mock('vue-router', () => ({
  useRoute: () => ({ path: mockRoutePath.value }),
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key,
    }),
  }
})

vi.mock('@/api/admin/system', () => ({
  checkUpdates: vi.fn(),
  default: {
    checkUpdates: vi.fn(),
  },
}))

vi.mock('@/api/auth', () => ({
  getPublicSettings: vi.fn(),
}))

function seedContactChannel() {
  const appStore = useAppStore()
  appStore.cachedPublicSettings = {
    registration_enabled: true,
    email_verify_enabled: false,
    registration_email_suffix_whitelist: [],
    promo_code_enabled: false,
    password_reset_enabled: false,
    invitation_code_enabled: false,
    referral_enabled: false,
    turnstile_enabled: false,
    turnstile_site_key: '',
    site_name: 'NanaFox Router',
    site_logo: '',
    site_subtitle: '',
    api_base_url: '',
    contact_info: '',
    doc_url: '',
    home_content: '',
    hide_ccs_import_button: false,
    purchase_subscription_enabled: true,
    purchase_subscription_url: '',
    custom_menu_items: [],
    custom_endpoints: [],
    linuxdo_oauth_enabled: false,
    oidc_oauth_enabled: false,
    oidc_oauth_provider_name: '',
    sora_client_enabled: false,
    backend_mode_enabled: false,
    version: 'test',
    contact_channels: [
      {
        type: 'wechat_group',
        label: '微信群',
        qr_image: 'https://example.test/qr.png',
        description: '',
        extra_info: '',
        enabled: true,
        priority: 1,
      },
    ],
  }
}

describe('FloatingContactButton route exclusions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    mockRoutePath.value = '/'
    seedContactChannel()
  })

  it('renders on normal user pages', () => {
    mockRoutePath.value = '/dashboard'

    const wrapper = mount(FloatingContactButton)

    expect(wrapper.text()).toContain('contact.label')
  })

  it.each(['/purchase', '/purchase/history', '/payment/result', '/payment/stripe'])(
    'does not render on payment flow route %s',
    (path) => {
      mockRoutePath.value = path

      const wrapper = mount(FloatingContactButton)

      expect(wrapper.text()).toBe('')
    },
  )
})
