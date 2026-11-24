import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import DocsSignupCta from '@/components/docs/DocsSignupCta.vue'

const state = vi.hoisted(() => ({
  isAuthenticated: false,
  settings: null as Record<string, unknown> | null
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string, params?: Record<string, unknown>) => (params ? `${key}:${JSON.stringify(params)}` : key) })
}))
vi.mock('@/stores', () => ({
  useAppStore: () => ({ cachedPublicSettings: state.settings }),
  useAuthStore: () => ({ isAuthenticated: state.isAuthenticated })
}))

function mountCta() {
  return mount(DocsSignupCta, {
    global: { stubs: { RouterLink: { props: ['to'], template: '<a :data-to="to"><slot /></a>' } } }
  })
}

describe('DocsSignupCta', () => {
  beforeEach(() => {
    state.isAuthenticated = false
    state.settings = { registration_enabled: true, default_balance: 5 }
  })

  it('invites visitors to register with the signup bonus', () => {
    const wrapper = mountCta()
    expect(wrapper.text()).toContain('docs.signupCta.titleWithBonus')
    expect(wrapper.text()).toContain('$5')
    expect(wrapper.get('a').attributes('data-to')).toBe('/register')
  })

  it('falls back to a plain signup prompt without a bonus or settings', () => {
    state.settings = null
    const wrapper = mountCta()
    expect(wrapper.text()).toContain('docs.signupCta.title')
    expect(wrapper.text()).toContain('docs.signupCta.button')
    expect(wrapper.text()).not.toContain('$')
  })

  it('is hidden for signed-in users and when registration is closed', () => {
    state.isAuthenticated = true
    expect(mountCta().find('[data-testid="docs-signup-cta"]').exists()).toBe(false)

    state.isAuthenticated = false
    state.settings = { registration_enabled: false, default_balance: 5 }
    expect(mountCta().find('[data-testid="docs-signup-cta"]').exists()).toBe(false)
  })
})
