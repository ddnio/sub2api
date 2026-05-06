import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
}))

const locationState = vi.hoisted(() => ({
  current: { href: 'http://localhost/register' } as { href: string },
}))

vi.mock('vue-router', () => ({
  useRoute: () => routeState,
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, string>) => {
        if (key === 'auth.oidc.signIn') return `Continue with ${params?.providerName ?? ''}`.trim()
        if (key === 'auth.linuxdo.signIn') return 'Continue with LinuxDo'
        if (key === 'auth.oauthOrContinue') return 'or continue'
        return key
      },
    }),
  }
})

describe('OAuth affiliate start URLs', () => {
  beforeEach(() => {
    routeState.query = { redirect: '/dashboard', aff_code: 'ABCDEF123456' }
    locationState.current = { href: 'http://localhost/register' }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: locationState.current,
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('passes affiliate codes to LinuxDo OAuth start', async () => {
    const wrapper = mount(LinuxDoOAuthSection)

    await wrapper.get('button').trigger('click')

    expect(locationState.current.href).toContain('/api/v1/auth/oauth/linuxdo/start?redirect=%2Fdashboard')
    expect(locationState.current.href).toContain('aff_code=ABCDEF123456')
  })

  it('passes affiliate codes to OIDC OAuth start', async () => {
    const wrapper = mount(OidcOAuthSection)

    await wrapper.get('button').trigger('click')

    expect(locationState.current.href).toContain('/api/v1/auth/oauth/oidc/start?redirect=%2Fdashboard')
    expect(locationState.current.href).toContain('aff_code=ABCDEF123456')
  })
})
