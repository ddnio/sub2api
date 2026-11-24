import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LowBalanceBanner from '@/components/layout/LowBalanceBanner.vue'

const state = vi.hoisted(() => ({
  user: null as { role: string; balance: number } | null,
  settings: null as Record<string, unknown> | null,
  hasActiveSubscriptions: false,
  path: '/dashboard'
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key) })
}))
vi.mock('vue-router', () => ({ useRoute: () => ({ path: state.path }) }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ cachedPublicSettings: state.settings }) }))
vi.mock('@/stores/auth', () => ({ useAuthStore: () => ({ user: state.user }) }))
vi.mock('@/stores/subscriptions', () => ({
  useSubscriptionStore: () => ({ hasActiveSubscriptions: state.hasActiveSubscriptions })
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

describe('LowBalanceBanner', () => {
  beforeEach(() => {
    state.user = { role: 'user', balance: 0 }
    state.settings = { balance_low_notify_threshold: 1, balance_low_notify_recharge_url: `${window.location.origin}/purchase` }
    state.hasActiveSubscriptions = false
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

  it('only warns about an empty balance when the threshold is off', () => {
    state.user = { role: 'user', balance: 0.42 }
    state.settings = { balance_low_notify_threshold: 0 }
    expect(mountBanner().find('[data-testid="low-balance-banner"]').exists()).toBe(false)
  })

  it('stays hidden for healthy balances, admins, subscribers and on the recharge page', () => {
    state.user = { role: 'user', balance: 3 }
    expect(mountBanner().find('[data-testid="low-balance-banner"]').exists()).toBe(false)

    state.user = { role: 'admin', balance: 0 }
    expect(mountBanner().find('[data-testid="low-balance-banner"]').exists()).toBe(false)

    state.user = { role: 'user', balance: 0 }
    state.hasActiveSubscriptions = true
    expect(mountBanner().find('[data-testid="low-balance-banner"]').exists()).toBe(false)

    state.hasActiveSubscriptions = false
    state.path = '/purchase'
    expect(mountBanner().find('[data-testid="low-balance-banner"]').exists()).toBe(false)
  })

  it('uses an external recharge URL as a plain link and ignores unsafe ones', () => {
    state.settings = { balance_low_notify_recharge_url: 'https://pay.example.com/top-up' }
    expect(mountBanner().get('a').attributes('href')).toBe('https://pay.example.com/top-up')

    state.settings = { balance_low_notify_recharge_url: 'javascript:alert(1)' }
    expect(mountBanner().get('a').attributes('data-to')).toBe('/purchase')
  })
})
