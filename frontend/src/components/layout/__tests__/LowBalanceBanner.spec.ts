import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LowBalanceBanner from '@/components/layout/LowBalanceBanner.vue'

const state = vi.hoisted(() => ({
  user: null as { role: string; balance: number } | null,
  settings: null as Record<string, unknown> | null,
  settingsLoaded: true,
  subscriptionsLoaded: true,
  hasActiveSubscriptions: false,
  subscriptionFeature: true,
  fetchActiveSubscriptions: vi.fn(),
  path: '/dashboard'
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key) })
}))
vi.mock('vue-router', () => ({ useRoute: () => ({ path: state.path }) }))
vi.mock('@/stores', () => ({
  useAppStore: () => ({ cachedPublicSettings: state.settings, publicSettingsLoaded: state.settingsLoaded })
}))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: state.user }) }))
vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({
    loaded: state.subscriptionsLoaded,
    hasActiveSubscriptions: state.hasActiveSubscriptions,
    fetchActiveSubscriptions: state.fetchActiveSubscriptions
  })
}))
vi.mock('@/utils/featureFlags', () => ({
  FeatureFlags: { subscription: 'subscription' },
  isFeatureFlagEnabled: () => state.subscriptionFeature
}))

function mountBanner() {
  return mount(LowBalanceBanner, {
    global: {
      stubs: {
        Icon: true,
        RouterLink: { props: ['to'], template: '<a :data-to="to"><slot /></a>' }
      }
    }
  })
}

const bannerShown = () => mountBanner().find('[data-testid="low-balance-banner"]').exists()

describe('LowBalanceBanner', () => {
  beforeEach(() => {
    state.user = { role: 'user', balance: 0 }
    state.settings = {
      balance_low_notify_enabled: true,
      balance_low_notify_threshold: 1,
      balance_low_notify_recharge_url: `${window.location.origin}/purchase`
    }
    state.settingsLoaded = true
    state.subscriptionsLoaded = true
    state.hasActiveSubscriptions = false
    state.subscriptionFeature = true
    state.fetchActiveSubscriptions = vi.fn().mockResolvedValue([])
    state.path = '/dashboard'
  })

  it('shows the empty-balance notice with an in-app recharge link', () => {
    const wrapper = mountBanner()
    expect(wrapper.text()).toContain('firstPayNudges.balanceEmpty')
    expect(wrapper.get('a').attributes('data-to')).toBe('/purchase')
  })

  it('shows the low-balance notice below the configured threshold', () => {
    state.user = { role: 'user', balance: 0.42 }
    const wrapper = mountBanner()
    expect(wrapper.text()).toContain('firstPayNudges.balanceLow')
    expect(wrapper.text()).toContain('$0.42')
  })

  it('only warns about an empty balance when the threshold is off or reminders are disabled', () => {
    state.user = { role: 'user', balance: 0.42 }
    state.settings = { balance_low_notify_enabled: true, balance_low_notify_threshold: 0 }
    expect(bannerShown()).toBe(false)

    state.settings = { balance_low_notify_enabled: false, balance_low_notify_threshold: 5 }
    expect(bannerShown()).toBe(false)

    state.user = { role: 'user', balance: 0 }
    expect(bannerShown()).toBe(true)
  })

  it('waits for settings and subscriptions before deciding', () => {
    state.settingsLoaded = false
    expect(bannerShown()).toBe(false)

    state.settingsLoaded = true
    state.subscriptionsLoaded = false
    expect(bannerShown()).toBe(false)
    expect(state.fetchActiveSubscriptions).toHaveBeenCalled()

    state.subscriptionFeature = false
    expect(bannerShown()).toBe(true)
  })

  it('stays hidden for healthy balances, admins, subscribers and on the recharge page', () => {
    state.user = { role: 'user', balance: 3 }
    expect(bannerShown()).toBe(false)

    state.user = { role: 'admin', balance: 0 }
    expect(bannerShown()).toBe(false)

    state.user = { role: 'user', balance: 0 }
    state.hasActiveSubscriptions = true
    expect(bannerShown()).toBe(false)

    state.hasActiveSubscriptions = false
    state.path = '/purchase/'
    expect(bannerShown()).toBe(false)
  })

  it('opens an external recharge URL in a new tab and ignores invalid ones', () => {
    state.settings = { balance_low_notify_recharge_url: 'https://pay.example.com/top-up' }
    const link = mountBanner().get('a')
    expect(link.attributes('href')).toBe('https://pay.example.com/top-up')
    expect(link.attributes('rel')).toBe('noopener noreferrer')

    for (const bad of ['javascript:alert(1)', 'https://', '//evil.example.com']) {
      state.settings = { balance_low_notify_recharge_url: bad }
      expect(mountBanner().get('a').attributes('data-to')).toBe('/purchase')
    }
  })
})
