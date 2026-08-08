import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import LinuxDoOAuthSection from '@/components/auth/LinuxDoOAuthSection.vue'
import OidcOAuthSection from '@/components/auth/OidcOAuthSection.vue'
import { buildOAuthLoginStartURL, type OAuthLoginStart } from '@/api/auth'

const routeState = vi.hoisted(() => ({
  query: {} as Record<string, unknown>,
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
    window.sessionStorage.clear()
  })

  it('passes affiliate codes to LinuxDo OAuth start', async () => {
    const wrapper = mount(LinuxDoOAuthSection)

    await wrapper.get('button').trigger('click')

    const request = wrapper.emitted('start')?.[0]?.[0] as OAuthLoginStart
    const startURL = buildOAuthLoginStartURL(request)
    expect(startURL).toContain('/api/v1/auth/oauth/linuxdo/start?redirect=%2Fdashboard')
    expect(startURL).toContain('aff_code=ABCDEF123456')
  })

  it('passes affiliate codes to OIDC OAuth start', async () => {
    const wrapper = mount(OidcOAuthSection)

    await wrapper.get('button').trigger('click')

    const request = wrapper.emitted('start')?.[0]?.[0] as OAuthLoginStart
    const startURL = buildOAuthLoginStartURL(request)
    expect(startURL).toContain('/api/v1/auth/oauth/oidc/start?redirect=%2Fdashboard')
    expect(startURL).toContain('aff_code=ABCDEF123456')
  })
})
