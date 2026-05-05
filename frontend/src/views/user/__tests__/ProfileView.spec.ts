<<<<<<< HEAD
import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import ProfileView from '@/views/user/ProfileView.vue'
import type { PublicSettings, UserProfile } from '@/types'

const publicSettingsState = vi.hoisted(() => ({
  value: null as PublicSettings | null,
}))

const profileState = vi.hoisted(() => ({
  value: null as UserProfile | null,
}))

vi.mock('@/api', () => ({
  authAPI: {
    getPublicSettings: vi.fn(async () => publicSettingsState.value),
  },
  userAPI: {
    getProfile: vi.fn(async () => profileState.value),
  },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: profileState.value,
  }),
=======
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ProfileView from '@/views/user/ProfileView.vue'

const {
  fetchPublicSettingsMock,
  refreshUserMock,
  authState
} = vi.hoisted(() => ({
  fetchPublicSettingsMock: vi.fn(),
  refreshUserMock: vi.fn(),
  authState: {
    user: null as Record<string, unknown> | null,
    refreshUser: vi.fn()
  }
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authState
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    fetchPublicSettings: fetchPublicSettingsMock
  })
}))

vi.mock('@/utils/format', () => ({
  formatDate: () => 'April 2026'
>>>>>>> v0.1.116
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({
<<<<<<< HEAD
      t: (key: string) => key,
    }),
  }
})

function createPublicSettings(overrides: Partial<PublicSettings> = {}): PublicSettings {
  return {
    site_name: 'Sub2API',
    site_logo: '',
    site_subtitle: '',
    registration_enabled: true,
    email_verify_enabled: false,
    registration_email_suffix_whitelist: [],
    promo_code_enabled: false,
    invitation_code_enabled: false,
    turnstile_site_key: '',
    api_base_url: '',
    contact_info: '',
    doc_url: '',
    home_content: '',
    hide_ccs_import_button: false,
    purchase_subscription_enabled: false,
    purchase_subscription_url: '',
    table_default_page_size: 20,
    table_page_size_options: [10, 20, 50, 100],
    custom_menu_items: [],
    custom_endpoints: [],
    linuxdo_oauth_enabled: false,
    wechat_oauth_enabled: false,
    wechat_oauth_open_enabled: false,
    wechat_oauth_mp_enabled: false,
    oidc_oauth_enabled: false,
    oidc_oauth_provider_name: 'OIDC',
    sora_client_enabled: false,
    backend_mode_enabled: false,
    version: 'test',
    balance_low_notify_enabled: false,
    account_quota_notify_enabled: false,
    balance_low_notify_threshold: 0,
    ...overrides,
  }
}

function createProfile(): UserProfile {
  return {
    id: 8,
    username: 'alice',
    email: 'alice@example.com',
    role: 'user',
    balance: 0,
    concurrency: 1,
    status: 'active',
    allowed_groups: null,
    balance_notify_enabled: true,
    balance_notify_threshold: null,
    balance_notify_extra_emails: [],
    created_at: '2026-05-01T00:00:00Z',
    updated_at: '2026-05-01T00:00:00Z',
  }
}

async function mountProfileView() {
  const wrapper = mount(ProfileView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        StatCard: { template: '<div />' },
        ProfileInfoCard: { template: '<div />', props: ['user'] },
        ProfileEditForm: { template: '<div />', props: ['initialUsername'] },
        ProfileBalanceNotifyCard: { template: '<div />' },
        ProfileIdentityBindingsSection: defineComponent({
          name: 'ProfileIdentityBindingsSection',
          template: '<div />',
          props: ['user', 'linuxdoEnabled', 'wechatEnabled', 'oidcEnabled', 'oidcProviderName'],
        }),
        ProfilePasswordForm: { template: '<div />' },
        ProfileTotpCard: { template: '<div />' },
        Icon: { template: '<span />' },
      },
    },
  })
  await nextTick()
  await nextTick()
  return wrapper
}

describe('ProfileView', () => {
  beforeEach(() => {
    publicSettingsState.value = createPublicSettings()
    profileState.value = createProfile()
    Object.defineProperty(window.navigator, 'userAgent', {
      configurable: true,
      value: 'Mozilla/5.0',
    })
  })

  it('shows WeChat binding on desktop only when Open OAuth is configured', async () => {
    publicSettingsState.value = createPublicSettings({
      wechat_oauth_enabled: true,
      wechat_oauth_mp_enabled: true,
    })

    const wrapper = await mountProfileView()

    expect(wrapper.getComponent({ name: 'ProfileIdentityBindingsSection' }).props('wechatEnabled')).toBe(false)
  })

  it('shows WeChat binding in WeChat browser when MP OAuth is configured', async () => {
    Object.defineProperty(window.navigator, 'userAgent', {
      configurable: true,
      value: 'MicroMessenger',
    })
    publicSettingsState.value = createPublicSettings({
      wechat_oauth_enabled: true,
      wechat_oauth_mp_enabled: true,
    })

    const wrapper = await mountProfileView()

    expect(wrapper.getComponent({ name: 'ProfileIdentityBindingsSection' }).props('wechatEnabled')).toBe(true)
=======
      t: (key: string) => key
    })
  }
})

describe('ProfileView', () => {
  beforeEach(() => {
    refreshUserMock.mockReset()
    fetchPublicSettingsMock.mockReset()
    refreshUserMock.mockResolvedValue(undefined)
    authState.refreshUser = refreshUserMock
    authState.user = {
      id: 1,
      username: 'alice',
      email: 'alice@example.com',
      role: 'user',
      balance: 10,
      concurrency: 2,
      status: 'active',
      allowed_groups: null,
      balance_notify_enabled: true,
      balance_notify_threshold: null,
      balance_notify_extra_emails: [],
      created_at: '2026-04-20T00:00:00Z',
      updated_at: '2026-04-20T00:00:00Z'
    }
    fetchPublicSettingsMock.mockResolvedValue({
      contact_info: '',
      balance_low_notify_enabled: false,
      balance_low_notify_threshold: 0,
      linuxdo_oauth_enabled: true,
      wechat_oauth_enabled: true,
      wechat_oauth_open_enabled: true,
      wechat_oauth_mp_enabled: false,
      oidc_oauth_enabled: true,
      oidc_oauth_provider_name: 'OIDC'
    })
  })

  it('renders the simplified single-column profile shell without separate stat cards', async () => {
    const wrapper = mount(ProfileView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          StatCard: { template: '<div class="stat-card" />' },
          ProfileInfoCard: { template: '<div data-testid="profile-info-card" />' },
          ProfileBalanceNotifyCard: { template: '<div data-testid="profile-balance-notify-card" />' },
          ProfilePasswordForm: { template: '<div data-testid="profile-password-form" />' },
          ProfileTotpCard: { template: '<div data-testid="profile-totp-card" />' },
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.findAll('.stat-card')).toHaveLength(0)
    expect(wrapper.get('[data-testid="profile-shell"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="profile-shell"]').html()).toContain('profile-info-card')
    expect(wrapper.get('[data-testid="profile-shell"]').html()).toContain('profile-password-form')
    expect(wrapper.get('[data-testid="profile-shell"]').html()).toContain('profile-totp-card')
>>>>>>> v0.1.116
  })
})
